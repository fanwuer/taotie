package spider

import (
	"fmt"
	"io/ioutil"
	"taotie/core/util"
	"testing"
)

func TestParseList(t *testing.T) {
	raw, _ := ioutil.ReadFile("./3.html")
	m, err := ParseListOtherType(raw)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for _, v := range m {
		fmt.Printf("%#v\n", v)
	}
}

func TestParseDetail(t *testing.T) {
	fs, err := util.ListDir("/Users/zhujiang/Documents/jinhan/taotie/data/asin", ".html")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(fs)
	for _, v := range fs {
		raw, _ := ioutil.ReadFile(v)
		m, err := ParseDetail(raw)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if m.Title == "" || m.Title == "Sorry! Something went wrong!" {
			continue
		}
		//if m.Img != "" {
		//	continue
		//}

		if !m.IsStock{
			//continue
		}
		if m.SoldBy != "" {
			//fmt.Println(v, m.SoldBy,m.SoldById)
			//continue
		} else {
		}

		if m.Price!=0{
			continue
		}
		fmt.Printf("%#v,%#v\n", v, m)
	}
}
