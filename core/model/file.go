package model

import (
	"errors"
	"taotie/core/util"
	"time"
)

type File struct {
	Id             int64  `json:"id" xorm:"bigint pk autoincr"`
	Type           string `json:"type" xorm:"index"`
	Tag            string `json:"tag" xorm:"index"`
	UserId         int64  `json:"user_id" xorm:"bigint index"`
	UserName       string `json:"user_name" xorm:"index"`
	FileName       string `json:"file_name"`
	ReallyFileName string `json:"really_file_name"`
	HashCode       string `json:"hash_code" xorm:"unique"`
	Url            string `json:"url" xorm:"varchar(700)"`
	UrlHashCode    string `json:"url_hash_code" xorm:"unique"`
	Describe       string `json:"describe" xorm:"TEXT"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time,omitempty"`
	Status         int    `json:"status" xorm:"notnull default(0) comment('0 normal，1 hide but can use') TINYINT(1)"`
	StoreType      int    `json:"store_type" xorm:"notnull default(0) comment('0 local，1 oss') TINYINT(1)"`
	IsPicture      int    `json:"is_picture"`
	Size           int64  `json:"size"`
}

var FileSortName = []string{"=id", "-create_time", "-update_time", "=user_id", "=type", "=tag", "=store_type", "=status", "=size"}

func (f *File) Exist() (bool, error) {
	if f.Id == 0 && f.Url == "" {
		return false, errors.New("where is empty")
	}
	s := Rdb.Client.Table(f)
	s.Where("1=1")

	if f.Id != 0 {
		s.And("id=?", f.Id)
	}
	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		s.And("url_hash_code=?", h)
	}

	c, err := s.Where("is_picture=?", 1).Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (f *File) Get() (bool, error) {
	if f.Id == 0 && f.Url == "" && f.UrlHashCode == "" && f.HashCode == "" {
		return false, errors.New("where is empty")
	}

	if f.Url != "" {
		h, err := util.Sha256([]byte(f.Url))
		if err != nil {
			return false, err
		}
		f.UrlHashCode = h
		f.Url = ""
	}

	return Rdb.Client.Get(f)
}

func (f *File) Update(hide bool) (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := Rdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id)

	if hide {
		f.Status = 1
		s.Cols("status")
	}

	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	if f.Describe != "" {
		s.Cols("describe")
	}

	if f.Tag != "" {
		s.Cols("tag")
	}

	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")
	_, err := s.Update(f)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (f *File) UpdateStatus() (bool, error) {
	if f.Id == 0 {
		return false, errors.New("where is empty")
	}

	s := Rdb.Client.NewSession()
	defer s.Close()

	s.Where("id=?", f.Id).Cols("status")

	if f.UserId != 0 {
		s.And("user_id=?", f.UserId)
	}

	f.UpdateTime = time.Now().Unix()
	s.Cols("update_time")

	_, err := s.Update(f)
	if err != nil {
		return false, err
	}

	return true, nil
}
