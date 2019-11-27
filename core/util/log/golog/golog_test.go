// Date: 18-1-9

package golog

import (
	"testing"
	"fmt"
	"runtime"
	"os"
	"time"
	"sync"
)

func TestFatal(t *testing.T) {
	fmt.Println(Ldate, Ltime, Lmicroseconds)
}

func TestItoa(t *testing.T) {
	a := []byte("abc")
	itoa(&a, 106, 5)
	fmt.Println(string(a))
}

func TestNew(t *testing.T) {
	var s string = ""
	fmt.Println([]byte(s))
	a := []byte("abc")
	a = append(a, ""...)
	fmt.Println(len(a))
}

func TestFatalf(t *testing.T) {
	a2()
}

func a2() {
	// 0 表示自己
	// 1 表示上一层调用的
	// 越往上那么逐次加一
	_, file, line, ok := runtime.Caller(1)
	fmt.Println(file, line, ok)
}

func TestOutput(t *testing.T) {
	runtime.GOMAXPROCS(1)
	l := New(os.Stdout, "", Lshortfile)
	for k := 1; k < 1000; k++ {
		//go l.Output(1, "a")
		l.Output(1, "a")
	}

	time.Sleep(time.Duration(100) * time.Second)
}


func TestPanicln(t *testing.T) {
	m:=sync.Mutex{}
	m.Lock()
	go func(){
		time.Sleep(time.Duration(2)*time.Second)
		m.Unlock()
	}()
	fmt.Println("ddd")
	m.Lock()
}