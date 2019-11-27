package oss

import (
	"bytes"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Key struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyId     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	BucketName      string `yaml:"bucket_name"`
}

func SaveFile(K Key, ObjectName string, raw []byte) error {
	// create OSSClient instance
	client, err := oss.New(K.Endpoint, K.AccessKeyId, K.AccessKeySecret)
	if err != nil {
		return err
	}

	// take bucket
	bucket, err := client.Bucket(K.BucketName)
	if err != nil {
		return err
	}

	// put bucket
	err = bucket.PutObject(ObjectName, bytes.NewReader(raw))
	if err != nil {
		return err
	}

	return nil
}
