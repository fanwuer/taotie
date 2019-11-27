package model

import (
	"errors"
	"fmt"
	"time"
)

type Group struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	Name       string `json:"name" xorm:"varchar(100) notnull unique"`
	Describe   string `json:"describe" xorm:"TEXT"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time,omitempty"`
	ImagePath  string `json:"image_path" xorm:"varchar(700)"`
}

var GroupSortName = []string{"=id", "=name", "-create_time", "=update_time"}

type Resource struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	Name       string `json:"name"`
	Url        string `json:"url"`
	UrlHash    string `json:"url_hash" xorm:"unique"`
	Describe   string `json:"describe" xorm:"TEXT"`
	Admin      bool   `json:"admin"`
	CreateTime int64  `json:"create_time"`
}

var ResourceSortName = []string{"=id", "+create_time", "-name"}

type GroupResource struct {
	Id         int64 `json:"id" xorm:"bigint pk autoincr"`
	GroupId    int64 `json:"group_id index(gr)"`
	ResourceId int64 `json:"resource_id index(gr)"`
}

func (g *Group) GetById() (exist bool, err error) {
	if g.Id == 0 {
		return false, errors.New("where is empty")
	}
	exist, err = Rdb.Client.Get(g)
	return
}

func (g *Group) Update() error {
	if g.Id == 0 {
		return errors.New("where is empty")
	}

	g.UpdateTime = time.Now().Unix()
	_, err := Rdb.Client.Where("id=?", g.Id).Omit("id").Update(g)
	return err
}

func (g *Group) Exist() (bool, error) {
	if g.Id == 0 && g.Name == "" {
		return false, errors.New("where is empty")
	}

	s := Rdb.Client.Table(g)
	s.Where("1=1")

	if g.Id != 0 {
		s.And("id=?", g.Id)
	}
	if g.Name != "" {
		s.And("name=?", g.Name)
	}

	c, err := s.Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}

func (g *Group) Delete() error {
	if g.Id == 0 && g.Name == "" {
		return errors.New("where is empty")
	}

	_, err := Rdb.Client.Delete(g)
	return err
}

func (g *Group) Take() (bool, error) {
	ok, err := g.Exist()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return Rdb.Client.Get(g)

}

func (r *Resource) Get() (err error) {
	var exist bool
	exist, err = Rdb.Client.UseBool("admin").Get(r)
	if err != nil {
		return
	}
	if !exist {
		return fmt.Errorf("resource not found")
	}
	return
}

func (r *Resource) GetRaw() (bool, error) {
	return Rdb.Client.UseBool("admin").Get(r)
}

func (r *Resource) InsertOne() (err error) {
	_, err = Rdb.Client.InsertOne(r)
	if err != nil {
		return
	}
	return
}

func (gr *GroupResource) Exist() (bool, error) {
	if gr.Id == 0 && gr.GroupId == 0 && gr.ResourceId == 0 {
		return false, errors.New("where is empty")
	}

	s := Rdb.Client.Table(gr)
	s.Where("1=1")

	if gr.Id != 0 {
		s.And("id=?", gr.Id)
	}
	if gr.GroupId != 0 {
		s.And("group_id=?", gr.GroupId)
	}

	if gr.ResourceId != 0 {
		s.And("resource_id=?", gr.ResourceId)
	}
	c, err := s.Count()

	if c >= 1 {
		return true, nil
	}

	return false, err
}
