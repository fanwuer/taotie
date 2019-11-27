package util

import (
	"fmt"
	"testing"
	"time"
)

func TestGetSecondTimes(t *testing.T) {
	// Mon Jan 2 15:04:05 -0700 MST 2006
	layOut := "Mon, 2 Jan 2006 15:04:05 GMT"
	s := "Wed, 23 Oct 2019 02:10:13 GMT"
	tt, err := time.ParseInLocation(layOut, s, time.UTC)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(tt.String())

	ExpirationLayOut := "2006-01-02T15:04:05Z"
	s = "2019-10-23T06:43:28Z"
	tt, err = time.ParseInLocation(ExpirationLayOut, s, time.UTC)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(tt.String())
}
