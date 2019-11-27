package model

import "time"

// https://www.amazon.com/Best-Sellers-Appliances-Clothes-Washing-Machines/zgbs/appliances/13397491
type AwsCategoryTask struct {
	Id            int64  `json:"id" xorm:"bigint pk autoincr"`
	Name          string `json:"name"`
	Link          string `json:"link" xorm:"TEXT"`
	LinkHashCode  string `json:"-" xorm:"unique"`
	CreateTime    int64  `json:"create_time"`
	UpdateTime    int64  `json:"update_time,omitempty"`
	Remark        string `json:"remark" xorm:"TEXT"`
	Open          int64  `json:"open" xorm:"index"`
	Type          int64  `json:"type" xorm:"index"`
	CatchDetail   int64  `json:"catch_detail" xorm:"index"`
	PageNum       int64  `json:"page_num"`
	Status        int64  `json:"status" xorm:"index"`
	OrderNum      int64  `json:"order_num"`
	Tag           string `json:"tag" xorm:"index"`
	LastCatchTime int64  `json:"last_catch_time" xorm:"index"`
	Todo          string `json:"todo" xorm:"-"`
	Doing         string `json:"doing" xorm:"-"`
	Done          string `json:"done" xorm:"-"`
}

var AwsCategoryTaskSortName = []string{"-order_num", "-last_catch_time", "-create_time", "-tag", "-update_time"}

func (t *AwsCategoryTask) Exist() (bool, error) {
	return Rdb.Client.Cols("link_hash_code").Exist(t)
}

func (t *AwsCategoryTask) InsertOne() (int64, error) {
	t.CreateTime = time.Now().Unix()
	return Rdb.Client.InsertOne(t)
}

func (t *AwsCategoryTask) GetRaw() (bool, error) {
	return Rdb.Client.Get(t)
}

func (t *AwsCategoryTask) Update() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).MustCols("catch_detail", "open", "status").Update(t)
}

func (t *AwsCategoryTask) UpdateAll() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).Cols("update_time", "name", "remark", "tag", "order_num", "page_num", "catch_detail", "open", "status").Update(t)
}

type AwsAsinTask struct {
	Id            int64  `json:"id" xorm:"bigint pk autoincr"`
	Name          string `json:"name"`
	Asin          string `json:"asin" xorm:"unique"`
	CreateTime    int64  `json:"create_time" xorm:"index"`
	UpdateTime    int64  `json:"update_time,omitempty" xorm:"index"`
	Remark        string `json:"remark" xorm:"TEXT"`
	Open          int64  `json:"open" xorm:"index"`
	Status        int64  `json:"status" xorm:"index"`
	OrderNum      int64  `json:"order_num" xorm:"index"`
	Tag           string `json:"tag" xorm:"index"`
	LastCatchTime int64  `json:"last_catch_time" xorm:"index"`
	Todo          string `json:"todo" xorm:"-"`
	Doing         string `json:"doing" xorm:"-"`
	Done          string `json:"done" xorm:"-"`
}

var AwsAsinTaskSortName = []string{"-order_num", "-last_catch_time", "-create_time", "-update_time", "-tag"}

func (t *AwsAsinTask) Exist() (bool, error) {
	return Rdb.Client.Cols("asin").Exist(t)
}

func (t *AwsAsinTask) InsertOne() (int64, error) {
	t.CreateTime = time.Now().Unix()
	return Rdb.Client.InsertOne(t)
}

func (t *AwsAsinTask) GetRaw() (bool, error) {
	return Rdb.Client.Get(t)
}

func (t *AwsAsinTask) Update() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).MustCols("open", "status", "is404").Update(t)
}

func (t *AwsAsinTask) UpdateAll() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).Cols("update_time", "name", "remark", "tag", "order_num", "open", "status", "is404").Update(t)
}

type AwsAsin struct {
	Id               int64   `json:"id" xorm:"bigint pk autoincr"`
	CategoryTaskId   int64   `json:"category_task_id" xorm:"index"`   // 0 or others
	CategoryTaskType int64   `json:"category_task_type" xorm:"index"` // 0 1 2
	SmallRank        int64   `json:"small_rank" xorm:"index"`         // 1 will have
	IsDetail         int64   `json:"is_detail" xorm:"index"`          // detail
	Name             string  `json:"name" xorm:"TEXT"`
	Asin             string  `json:"asin" xorm:"index"`
	BigRank          int64   `json:"big_rank" xorm:"index"` // detail will have
	Img              string  `json:"img" xorm:"TEXT"`
	Price            float64 `json:"price" xorm:"index"`
	Score            float64 `json:"score" xorm:"index"`
	IsFba            int64   `json:"fba" xorm:"index"`
	IsAwsSold        int64   `json:"is_aws_sold" xorm:"index"`
	SoldBy           string  `json:"sold_by" xorm:"index"`
	SoldById         string  `json:"sold_by_id" xorm:"index"`
	IsPrime          int64   `json:"is_prime" xorm:"index"`
	Describe         string  `json:"describe" xorm:"TEXT"`
	RankDetail       string  `json:"rank_detail" xorm:"TEXT"`
	CreateTime       int64   `json:"create_time" xorm:"index"`
	UpdateTime       int64   `json:"update_time,omitempty"`
	Status           int64   `json:"status" xorm:"index"`
	Tag              string  `json:"tag" xorm:"index"`
	Reviews          int64   `json:"reviews"`
	Remark           string  `json:"remark" xorm:"TEXT"`
}

var AwsAsinSortName = []string{"-create_time", "=big_rank", "=small_rank", "=price", "=score", "=star", "-update_time", "-tag"}

func (t *AwsAsin) GetRaw() (bool, error) {
	return Rdb.Client.Get(t)
}

func (t *AwsAsin) Update() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).MustCols("status").Update(t)
}

type AwsAsinLib struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	Asin       string `json:"asin" xorm:"unique"`
	CreateTime int64  `json:"create_time" xorm:"index"`
	UpdateTime int64  `json:"update_time,omitempty" xorm:"index"`
	Times      int64  `json:"times" xorm:"index"`
	Tag        string `json:"tag" xorm:"index"`
	Remark     string `json:"remark" xorm:"TEXT"`
	Todo       string `json:"todo" xorm:"-"`
	Doing      string `json:"doing" xorm:"-"`
	Done       string `json:"done" xorm:"-"`
}

var AwsAsinLibSortName = []string{"-times", "-update_time", "-create_time", "-tag"}

func (t *AwsAsinLib) GetRaw() (bool, error) {
	return Rdb.Client.Get(t)
}

func (t *AwsAsinLib) InsertOne() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.CreateTime = time.Now().Unix()
	return s.InsertOne(t)
}

func (t *AwsAsinLib) Update() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).Update(t)
}

func (t *AwsAsinLib) Incr() (int64, error) {
	s := Rdb.Client.NewSession()
	defer s.Close()
	t.UpdateTime = time.Now().Unix()
	return s.ID(t.Id).Incr("times").Cols("times", "update_time").Update(t)
}

type AwsStatistics struct {
	Id         int64  `json:"id" xorm:"bigint pk autoincr"`
	CreateTime int64  `json:"create_time" xorm:"index"`
	UpdateTime int64  `json:"update_time,omitempty" xorm:"index"`
	Times      int64  `json:"times"`
	Today      string `json:"today" xorm:"index"`
	Name       string `json:"name"`
	Type       int64  `json:"type" xorm:"index"`
}

var AwsStatisticsSortName = []string{"-create_time", "-update_time", "-today"}

func (h *AwsStatistics) GetRaw() (bool, error) {
	return Rdb.Client.Get(h)
}

func (h *AwsStatistics) InsertOne() (int64, error) {
	h.CreateTime = time.Now().Unix()
	return Rdb.Client.InsertOne(h)
}

func (h *AwsStatistics) UpdateCount() (int64, error) {
	h.UpdateTime = time.Now().Unix()
	return Rdb.Client.ID(h.Id).Cols("update_time", "times").Update(h)
}
