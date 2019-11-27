package oss

import (
	"fmt"
	"testing"
)

func TestSaveFile(t *testing.T) {
	k := Key{
		"oss-cn-qingdao.aliyuncs.com",
		"LTAItzR6DHxzgBTH",
		"kKEQ6mrNn6CJm8YjRVwtbBzvjTAynt",
		"syoss",
	}
	err := SaveFile(k, "jj/xx/afsaf.jpg", []byte("ddddd"))
	fmt.Printf("%#v", err)
}
