package util

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strings"
)

// GetGUID
func GetGUID() (valueGUID string) {
	objID := uuid.NewV4()
	objIdStr := objID.String()
	objIdStr = strings.Replace(objIdStr, "-", "", -1)
	valueGUID = objIdStr
	return valueGUID
}

// sha256 256 bit
func Sha256(raw []byte) (string, error) {
	h := sha256.New()
	num, err := h.Write(raw)
	if err != nil {
		return "", err
	}
	if num == 0 {
		return "", errors.New("num 0")
	}
	data := h.Sum([]byte(""))
	return fmt.Sprintf("%x", data), nil
}

func Md5(raw []byte) (string, error) {
	h := md5.New()
	num, err := h.Write(raw)
	if err != nil {
		return "", err
	}
	if num == 0 {
		return "", errors.New("num 0")
	}
	data := h.Sum([]byte(""))
	return fmt.Sprintf("%x", data), nil
}
