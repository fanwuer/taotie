package spider

import (
	"github.com/hunterhug/marmot/miner"
	"strings"
	"sync"
	"taotie/core/flog"
)

var (
	Spiders = &_Spider{ws: make(map[string]*miner.Worker)}
)

func init() {
	miner.DefaultTimeOut = MinerDefaultTimeOut
}

type _Spider struct {
	mux sync.RWMutex
	ws  map[string]*miner.Worker
}

func (sb *_Spider) Get(name string) (b *miner.Worker, ok bool) {
	sb.mux.RLock()
	b, ok = sb.ws[name]
	sb.mux.RUnlock()
	return
}

func (sb *_Spider) Set(name string, b *miner.Worker) {
	sb.mux.Lock()
	sb.ws[name] = b
	sb.mux.Unlock()
	return
}

func (sb *_Spider) Delete(name string) {
	sb.mux.Lock()
	delete(sb.ws, name)
	sb.mux.Unlock()
	return
}

func Download(ip string, url string) ([]byte, error) {
	browser, ok := Spiders.Get(ip)
	if ok {
		browser.Url = url
		content, err := browser.Get()
		flog.Log.Debugf("download url:%s,status:%d,ip:%s", url, browser.UrlStatuscode, ip)
		return content, err
	} else {
		browser, _ := miner.New(ip)
		browser.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		browser.Header.Set("Accept-Language", "en-US;q=0.8,en;q=0.5")
		browser.Header.Set("Connection", "keep-alive")
		if strings.Contains(url, "www.amazon.co.jp") {
			browser.Header.Set("Host", "www.amazon.co.jp")
		} else if strings.Contains(url, "www.amazon.de") {
			browser.Header.Set("Host", "www.amazon.de")
		} else if strings.Contains(url, "www.amazon.co.uk") {
			browser.Header.Set("Host", "www.amazon.co.uk")
		} else {
			browser.Header.Set("Host", "www.amazon.com")
		}
		browser.Header.Set("Upgrade-Insecure-Requests", "1")
		browser.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.71 Safari/537.36")
		browser.Url = url
		Spiders.Set(ip, browser)
		content, err := browser.Get()
		flog.Log.Debugf("download url:%s,status:%d,ip:%s", url, browser.UrlStatuscode, ip)
		return content, err
	}
}
