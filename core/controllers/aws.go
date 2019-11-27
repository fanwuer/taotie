package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"math"
	"strings"
	"taotie/core/flog"
	"taotie/core/model"
	"taotie/core/spider"
	"taotie/core/util"
)

type AwsAddCategoryTaskRequest struct {
	Name        string `json:"name" validate:"required"`
	Link        string `json:"link" validate:"required"`
	Remark      string `json:"remark"`
	Open        int64  `json:"open" validate:"oneof=0 1"`
	Type        int64  `json:"type" validate:"oneof=1 2"`
	CatchDetail int64  `json:"catch_detail" validate:"oneof=0 1"`
	PageNum     int64  `json:"page_num" validate:"oneof=1 2 3 4 5"`
	OrderNum    int64  `json:"order_num"`
	Tag         string `json:"tag"`
}

func AwsAddCategoryTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsAddCategoryTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsAddCategoryTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// https://www.amazon.com/Best-Sellers-Appliances-Clothes-Washing-Machines/zgbs/appliances/13397491
	// https://www.amazon.com/Best-Sellers/zgbs/fashion/ref=zg_bs_nav_0
	// https://www.amazon.com/s?me=A38PDDQNY9S8ES&marketplaceID=ATVPDKIKX0DER
	// https://www.amazon.com/s?me=A3K0O9OTHV8P5O
	if req.Type == 1 {
		req.Link = strings.TrimSpace(strings.Split(req.Link, "?")[0])
		if !strings.HasPrefix(req.Link, "https://www.amazon.com/Best-Sellers") {
			resp.Error = Error(ParasError, "prefix should https://www.amazon.com/Best-Sellers")
			return
		}
		if strings.Contains(req.Link, "/ref=zg_bs") {
			req.Link = strings.Split(req.Link, "/ref=zg_bs")[0]
		}
	} else {
		if !strings.HasPrefix(req.Link, "https://www.amazon.com/s?me=") {
			resp.Error = Error(ParasError, "prefix should https://www.amazon.com/s?me=")
			return
		}

		req.Link = strings.Split(strings.TrimSpace(req.Link), "&")[0]
	}

	task := new(model.AwsCategoryTask)
	task.LinkHashCode, _ = util.Md5([]byte(req.Link))
	ok, err := task.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsAddCategoryTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if ok {
		if task.Status == 0 {
			flog.Log.Errorf("AwsAddCategoryTask err:%s exist", task.Link)
			resp.Error = Error(DbRepeat, "")
			return
		}
		task.Name = req.Name
		task.Remark = req.Remark
		task.Tag = req.Tag
		task.OrderNum = req.OrderNum
		task.PageNum = req.PageNum
		task.Open = req.Open
		task.CatchDetail = req.CatchDetail
		task.Status = 0
		_, err = task.UpdateAll()
		if err != nil {
			flog.Log.Errorf("AwsAddCategoryTask err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		resp.Flag = true
		resp.Data = task.Id
		return
	}

	task.Link = req.Link
	task.Name = req.Name
	task.Type = req.Type
	task.Remark = req.Remark
	task.Tag = req.Tag
	task.OrderNum = req.OrderNum
	task.PageNum = req.PageNum
	task.Open = req.Open
	task.CatchDetail = req.CatchDetail
	_, err = task.InsertOne()
	if err != nil {
		flog.Log.Errorf("AwsAddCategoryTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = task.Id
}

type AwsUpdateCategoryTaskRequest struct {
	Id          int64  `json:"id" validate:"required"`
	Name        string `json:"name"`
	Remark      string `json:"remark"`
	Open        int64  `json:"open" validate:"oneof=0 1"`
	CatchDetail int64  `json:"catch_detail" validate:"oneof=0 1"`
	PageNum     int64  `json:"page_num" validate:"oneof=1 2 3 4 5"`
	OrderNum    int64  `json:"order_num"`
	Tag         string `json:"tag"`
	Status      int64  `json:"status" validate:"oneof=0 1"`
}

func AwsUpdateCategoryTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsUpdateCategoryTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsUpdateCategoryTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	t := new(model.AwsCategoryTask)
	t.Id = req.Id
	ok, err := t.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsUpdateCategoryTask err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("AwsUpdateCategoryTask err: %d not found", req.Id)
		resp.Error = Error(DbNotFound, "")
		return
	}

	newT := new(model.AwsCategoryTask)
	newT.Id = req.Id
	if req.Name != "" && req.Name != t.Name {
		newT.Name = req.Name
	}
	if req.Remark != "" && req.Remark != t.Remark {
		newT.Remark = req.Remark
	}

	if req.Tag != "" && req.Tag != t.Tag {
		newT.Tag = req.Tag
	}
	newT.CatchDetail = req.CatchDetail
	newT.Open = req.Open
	newT.PageNum = req.PageNum
	newT.OrderNum = req.OrderNum
	newT.Status = req.Status
	_, err = newT.Update()
	if err != nil {
		flog.Log.Errorf("AwsUpdateCategoryTask err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type AwsListCategoryTaskRequest struct {
	Id                 int64    `json:"id"`
	Open               int64    `json:"open" validate:"oneof=-1 0 1"`
	CatchDetail        int64    `json:"catch_detail" validate:"oneof=-1 0 1"`
	PageNum            int64    `json:"page_num" validate:"oneof=-1 1 2 3 4 5"`
	Tag                string   `json:"tag"`
	Status             int64    `json:"status" validate:"oneof=0 1"`
	Link               string   `json:"link"`
	Type               int64    `json:"type" validate:"oneof=-1 1 2"`
	CreateTimeBegin    int64    `json:"create_time_begin"`
	CreateTimeEnd      int64    `json:"create_time_end"`
	UpdateTimeBegin    int64    `json:"update_time_begin"`
	UpdateTimeEnd      int64    `json:"update_time_end"`
	LastCatchTimeBegin int64    `json:"last_catch_time_begin"`
	LastCatchTimeEnd   int64    `json:"last_catch_time_end"`
	Sort               []string `json:"sort"`
	PageHelp
}

type AwsListCategoryTaskResponse struct {
	Tasks []model.AwsCategoryTask `json:"task"`
	PageHelp
}

func AwsListCategoryTask(c *gin.Context) {
	resp := new(Resp)

	respResult := new(AwsListCategoryTaskResponse)
	req := new(AwsListCategoryTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsListCategoryTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.Rdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.AwsCategoryTask)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}
	if req.Open != -1 {
		session.And("open=?", req.Open)
	}

	if req.CatchDetail != -1 {
		session.And("catch_detail=?", req.Status)
	}

	if req.PageNum != -1 {
		session.And("page_num=?", req.PageNum)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	session.And("status=?", req.Status)

	if req.Link != "" {
		code, _ := util.Md5([]byte(req.Link))
		session.And("link_hash_code=?", code)
	}

	if req.Type != -1 {
		session.And("type=?", req.Type)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		if req.UpdateTimeBegin == 0 {
			session.And("update_time>?", 0)
		}
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.LastCatchTimeBegin > 0 {
		session.And("last_catch_time>=?", req.LastCatchTimeBegin)
	}

	if req.LastCatchTimeEnd > 0 {
		if req.LastCatchTimeBegin == 0 {
			session.And("last_catch_time>?", 0)
		}
		session.And("last_catch_time<?", req.LastCatchTimeEnd)
	}

	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("AwsListCategoryTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	tasks := make([]model.AwsCategoryTask, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		p.build(session, req.Sort, model.AwsCategoryTaskSortName)
		err = session.Find(&tasks)
		if err != nil {
			flog.Log.Errorf("AwsListCategoryTask err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	for k, v := range tasks {
		idStr := fmt.Sprintf("%d", v.Id)
		tasks[k].Todo, _, _ = spider.GetHashPool(spider.AwsCategoryHashPoolToDoName, idStr)
		tasks[k].Done, _, _ = spider.GetHashPool(spider.AwsCategoryHashPoolDoneName, idStr)
		tasks[k].Doing, _, _ = spider.GetHashPool(spider.AwsCategoryHashPoolDoingName, idStr)
	}
	respResult.Tasks = tasks
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type AwsAddAsinTaskRequest struct {
	Name     string `json:"name" validate:"required"`
	Asin     string `json:"asin" validate:"required"`
	Remark   string `json:"remark"`
	Open     int64  `json:"open" validate:"oneof=0 1"`
	OrderNum int64  `json:"order_num"`
	Tag      string `json:"tag"`
}

func AwsAddAsinTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsAddAsinTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsAddAsinTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	task := new(model.AwsAsinTask)
	task.Asin = req.Asin
	ok, err := task.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsAddAsinTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if ok {
		if task.Status == 0 {
			flog.Log.Errorf("AwsAddAsinTask err:%s exist", task.Asin)
			resp.Error = Error(DbRepeat, "")
			return
		}

		task.Status = 0
		task.Name = req.Name
		task.Remark = req.Remark
		task.Tag = req.Tag
		task.OrderNum = req.OrderNum
		task.Open = req.Open
		_, err = task.UpdateAll()
		if err != nil {
			flog.Log.Errorf("AwsAddAsinTask err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		resp.Flag = true
		resp.Data = task.Id
		return
	}

	task.Name = req.Name
	task.Asin = req.Asin
	task.Remark = req.Remark
	task.Tag = req.Tag
	task.OrderNum = req.OrderNum
	task.Open = req.Open
	_, err = task.InsertOne()
	if err != nil {
		flog.Log.Errorf("AwsAddAsinTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
	resp.Data = task.Id
}

type AwsUpdateAsinTaskRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required"`
	Remark   string `json:"remark"`
	Open     int64  `json:"open" validate:"oneof=0 1"`
	OrderNum int64  `json:"order_num"`
	Tag      string `json:"tag"`
	Status   int64  `json:"status" validate:"oneof=0 1"`
}

func AwsUpdateAsinTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsUpdateAsinTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	t := new(model.AwsAsinTask)
	t.Id = req.Id
	ok, err := t.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinTask err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("AwsUpdateAsinTask err: %d not found", req.Id)
		resp.Error = Error(DbNotFound, "")
		return
	}

	newT := new(model.AwsAsinTask)
	newT.Id = req.Id
	if req.Name != "" && req.Name != t.Name {
		newT.Name = req.Name
	}
	if req.Remark != "" && req.Remark != t.Remark {
		newT.Remark = req.Remark
	}

	if req.Tag != "" && req.Tag != t.Tag {
		newT.Tag = req.Tag
	}

	newT.Open = req.Open
	newT.OrderNum = req.OrderNum
	newT.Status = req.Status
	_, err = newT.Update()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinTask err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type AwsListAsinTaskRequest struct {
	Id                 int64    `json:"id"`
	Asin               string   `json:"asin"`
	Open               int64    `json:"open" validate:"oneof=-1 0 1"`
	Tag                string   `json:"tag"`
	Status             int64    `json:"status" validate:"oneof=-1 0 1"`
	CreateTimeBegin    int64    `json:"create_time_begin"`
	CreateTimeEnd      int64    `json:"create_time_end"`
	UpdateTimeBegin    int64    `json:"update_time_begin"`
	UpdateTimeEnd      int64    `json:"update_time_end"`
	LastCatchTimeBegin int64    `json:"last_catch_time_begin"`
	LastCatchTimeEnd   int64    `json:"last_catch_time_end"`
	Sort               []string `json:"sort"`
	PageHelp
}

type AwsListAsinTaskResponse struct {
	Tasks []model.AwsAsinTask `json:"task"`
	PageHelp
}

func AwsListAsinTask(c *gin.Context) {
	resp := new(Resp)

	respResult := new(AwsListAsinTaskResponse)
	req := new(AwsListAsinTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsListAsinTask err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.Rdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.AwsAsinTask)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}
	if req.Open != -1 {
		session.And("open=?", req.Open)
	}

	if req.Asin != "" {
		session.And("asin=?", req.Asin)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	session.And("status=?", req.Status)

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		if req.UpdateTimeBegin == 0 {
			session.And("update_time>?", 0)
		}
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.LastCatchTimeBegin > 0 {
		session.And("last_catch_time>=?", req.LastCatchTimeBegin)
	}

	if req.LastCatchTimeEnd > 0 {
		if req.LastCatchTimeBegin == 0 {
			session.And("last_catch_time>?", 0)
		}
		session.And("last_catch_time<?", req.LastCatchTimeEnd)
	}

	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("AwsListAsinTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	tasks := make([]model.AwsAsinTask, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		p.build(session, req.Sort, model.AwsAsinTaskSortName)
		err = session.Find(&tasks)
		if err != nil {
			flog.Log.Errorf("AwsListAsinTask err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	for k, v := range tasks {
		idStr := v.Asin
		tasks[k].Todo, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolToDoName, idStr)
		tasks[k].Done, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolDoneName, idStr)
		tasks[k].Doing, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolDoingName, idStr)
	}

	respResult.Tasks = tasks
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type AwsListAsinLibRequest struct {
	Id              int64    `json:"id"`
	Asin            string   `json:"asin"`
	Tag             string   `json:"tag"`
	TimesBegin      int64    `json:"times_begin"`
	TimesEnd        int64    `json:"times_end"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	Sort            []string `json:"sort"`
	PageHelp
}

type AwsListAsinLibResponse struct {
	Asin []model.AwsAsinLib `json:"asin"`
	PageHelp
}

func AwsListAsinLib(c *gin.Context) {
	resp := new(Resp)

	respResult := new(AwsListAsinLibResponse)
	req := new(AwsListAsinLibRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsListAsinLib err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.Rdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.AwsAsinLib)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.Asin != "" {
		session.And("asin=?", req.Asin)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		if req.UpdateTimeBegin == 0 {
			session.And("update_time>?", 0)
		}
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.TimesBegin > 0 {
		session.And("times>=?", req.TimesBegin)
	}

	if req.TimesEnd > 0 {
		session.And("times<?", req.TimesEnd)
	}

	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("AwsListAsinLib err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	asin := make([]model.AwsAsinLib, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		p.build(session, req.Sort, model.AwsAsinLibSortName)
		err = session.Find(&asin)
		if err != nil {
			flog.Log.Errorf("AwsListAsinLib err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	for k, v := range asin {
		idStr := v.Asin
		asin[k].Todo, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolToDoName, idStr)
		asin[k].Done, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolDoneName, idStr)
		asin[k].Doing, _, _ = spider.GetHashPool(spider.AwsAsinHashPoolDoingName, idStr)
	}

	respResult.Asin = asin
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type AwsUpdateAsinLibRequest struct {
	Id     int64  `json:"id" validate:"required"`
	Tag    string `json:"tag"`
	Remark string `json:"remark"`
}

func AwsUpdateAsinLib(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsUpdateAsinLibRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinLib err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	t := new(model.AwsAsinLib)
	t.Id = req.Id
	ok, err := t.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinLib err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("AwsUpdateAsinLib err: %d not found", req.Id)
		resp.Error = Error(DbNotFound, "")
		return
	}

	newT := new(model.AwsAsinLib)
	newT.Id = req.Id
	if req.Remark != "" && req.Remark != t.Remark {
		newT.Remark = req.Remark
	}

	if req.Tag != "" && req.Tag != t.Tag {
		newT.Tag = req.Tag
	}

	_, err = newT.Update()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinLib err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type AwsListAsinDetailRequest struct {
	Id               int64    `json:"id"`
	Asin             string   `json:"asin"`
	Tag              string   `json:"tag"`
	Status           int64    `json:"status" validate:"oneof=0 1"`
	CreateTimeBegin  int64    `json:"create_time_begin"`
	CreateTimeEnd    int64    `json:"create_time_end"`
	UpdateTimeBegin  int64    `json:"update_time_begin"`
	UpdateTimeEnd    int64    `json:"update_time_end"`
	CategoryTaskId   int64    `json:"category_task_id"`
	CategoryTaskType int64    `json:"category_task_type" validate:"oneof=-1 0 1 2"`
	IsDetail         int64    `json:"is_detail" validate:"oneof=-1 0 1"`
	BigRankBegin     int64    `json:"big_rank_begin"`
	BigRankEnd       int64    `json:"big_rank_end"`
	SmallRankBegin   int64    `json:"small_rank_begin"`
	SmallRankEnd     int64    `json:"small_rank_end"`
	PriceBegin       float64  `json:"price_begin"`
	PriceEnd         float64  `json:"price_end"`
	ScoreBegin       float64  `json:"score_begin"`
	ScoreEnd         float64  `json:"score_end"`
	ReviewsBegin     int64    `json:"reviews_begin"`
	ReviewsEnd       int64    `json:"reviews_end"`
	IsFba            int64    `json:"is_fba" validate:"oneof=-1 0 1"`
	IsAwsSold        int64    `json:"is_aws_sold" validate:"oneof=-1 0 1"`
	IsPrime          int64    `json:"is_prime" validate:"oneof=-1 0 1"`
	SoldBy           string   `json:"sold_by"`
	SoldById         string   `json:"sold_by_id"`
	Sort             []string `json:"sort"`
	PageHelp
}

type AwsListAsinDetailResponse struct {
	Asin []model.AwsAsin `json:"asin"`
	PageHelp
}

func AwsListAsinDetail(c *gin.Context) {
	resp := new(Resp)

	respResult := new(AwsListAsinDetailResponse)
	req := new(AwsListAsinDetailRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsListAsinDetail err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.Rdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.AwsAsin)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.Asin != "" {
		session.And("asin=?", req.Asin)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	session.And("status=?", req.Status)

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		if req.UpdateTimeBegin == 0 {
			session.And("update_time>?", 0)
		}
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.CategoryTaskId != 0 {
		session.And("category_task_id=?", req.CategoryTaskId)
	}

	if req.CategoryTaskType != -1 {
		session.And("category_task_type=?", req.CategoryTaskType)
	}

	if req.IsDetail != -1 {
		session.And("is_detail=?", req.IsDetail)
	}

	if req.BigRankBegin > 0 {
		session.And("big_rank>=?", req.BigRankBegin)
	}

	if req.BigRankEnd > 0 {
		if req.BigRankBegin == 0 {
			session.And("big_rank>?", 0)
		}
		session.And("big_rank<?", req.BigRankEnd)
	}

	if req.SmallRankBegin > 0 {
		session.And("small_rank>=?", req.SmallRankBegin)
	}

	if req.SmallRankEnd > 0 {
		if req.SmallRankBegin == 0 {
			session.And("small_rank>?", 0)
		}
		session.And("small_rank<?", req.SmallRankEnd)
	}

	if req.PriceBegin > 0 {
		session.And("price>=?", req.PriceBegin)
	}

	if req.PriceEnd > 0 {
		session.And("price<?", req.PriceEnd)
	}

	if req.ScoreBegin > 0 {
		session.And("score>=?", req.ScoreBegin)
	}

	if req.ScoreEnd > 0 {
		session.And("score<?", req.ScoreEnd)
	}

	if req.ReviewsBegin > 0 {
		session.And("reviews>=?", req.ReviewsBegin)
	}

	if req.ReviewsEnd > 0 {
		session.And("reviews<?", req.ReviewsEnd)
	}

	if req.IsFba != -1 {
		session.And("is_fba=?", req.IsFba)
	}

	if req.IsAwsSold != -1 {
		session.And("is_aws_sold=?", req.IsAwsSold)
	}

	if req.IsPrime != -1 {
		session.And("is_prime=?", req.IsPrime)
	}

	if req.SoldBy != "" {
		session.And("sold_by=?", req.SoldBy)
	}

	if req.SoldById != "" {
		session.And("sold_by_id=?", req.SoldById)
	}

	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("AwsListAsinDetail err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	asin := make([]model.AwsAsin, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		p.build(session, req.Sort, model.AwsAsinSortName)
		err = session.Find(&asin)
		if err != nil {
			flog.Log.Errorf("AwsListAsinDetail err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	respResult.Asin = asin
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type AwsUpdateAsinDetailRequest struct {
	Id     int64  `json:"id" validate:"required"`
	Remark string `json:"remark"`
	Tag    string `json:"tag"`
	Status int64  `json:"status" validate:"oneof=0 1"`
}

func AwsUpdateAsinDetail(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsUpdateAsinDetailRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinDetail err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	t := new(model.AwsAsin)
	t.Id = req.Id
	ok, err := t.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinDetail err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("AwsUpdateAsinDetail err: %d not found", req.Id)
		resp.Error = Error(DbNotFound, "")
		return
	}

	newT := new(model.AwsAsin)
	newT.Id = req.Id
	if req.Remark != "" && req.Remark != t.Remark {
		newT.Remark = req.Remark
	}

	if req.Tag != "" && req.Tag != t.Tag {
		newT.Tag = req.Tag
	}

	newT.Status = req.Status
	_, err = newT.Update()
	if err != nil {
		flog.Log.Errorf("AwsUpdateAsinDetail err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Flag = true
}

type AwsListStatisticsRequest struct {
	Id              int64    `json:"id"`
	Type            int64    `json:"type" validate:"oneof=-1 0 1 2 3"`
	Today           string   `json:"today"`
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	Sort            []string `json:"sort"`
	PageHelp
}

type AwsListStatisticsResponse struct {
	Statistics []model.AwsStatistics `json:"statistics"`
	PageHelp
}

func AwsListStatistics(c *gin.Context) {
	resp := new(Resp)

	respResult := new(AwsListStatisticsResponse)
	req := new(AwsListStatisticsRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		flog.Log.Errorf("AwsListStatistics err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.Rdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.AwsStatistics)).Where("1=1")

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.Type != -1 {
		session.And("type=?", req.Type)
	}

	if req.Today != "" {
		session.And("today=?", req.Today)
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		if req.UpdateTimeBegin == 0 {
			session.And("update_time>?", 0)
		}
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		flog.Log.Errorf("AwsListAsinLib err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	statistics := make([]model.AwsStatistics, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		p.build(session, req.Sort, model.AwsStatisticsSortName)
		err = session.Find(&statistics)
		if err != nil {
			flog.Log.Errorf("AwsListStatistics err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	respResult.Statistics = statistics
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type AwsRunCategoryTaskRequest struct {
	Id int64 `json:"id" validate:"required"`
}

func AwsRunCategoryTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsRunCategoryTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Id == 0 {
		flog.Log.Errorf("AwsRunCategoryTask err:%s", "id is empty")
		resp.Error = Error(ParasError, "id is empty")
		return
	}
	t := new(model.AwsCategoryTask)
	t.Id = req.Id
	ok, err := t.GetRaw()
	if err != nil {
		flog.Log.Errorf("AwsRunCategoryTask err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !ok {
		flog.Log.Errorf("AwsRunCategoryTask err:%s", "not found")
		resp.Error = Error(DbNotFound, "")
		return
	}
	err = spider.AwsOneCategorySentToPoolRightNow(*t)
	if err != nil {
		flog.Log.Errorf("AwsRunCategoryTask err:%s", err.Error())
		resp.Error = Error(SpiderBusyError, err.Error())
		return
	}

	resp.Flag = true
	return
}

type AwsRunAsinTaskRequest struct {
	Id   int64  `json:"id"`
	Asin string `json:"asin"`
}

func AwsRunAsinTask(c *gin.Context) {
	resp := new(Resp)
	req := new(AwsRunAsinTaskRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Id == 0 && req.Asin == "" {
		flog.Log.Errorf("AwsRunAsinTask err:%s", "id and asin is empty")
		resp.Error = Error(ParasError, "id and asin is empty")
		return
	}
	t := new(model.AwsAsinTask)
	t.Id = req.Id

	if req.Id != 0 {
		ok, err := t.GetRaw()
		if err != nil {
			flog.Log.Errorf("AwsRunAsinTask err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		if !ok {
			flog.Log.Errorf("AwsRunAsinTask err:%s", "not found")
			resp.Error = Error(DbNotFound, "")
			return
		}
	} else {
		t.Asin = req.Asin
	}

	err := spider.AwsOneAsinSentToPoolRightNow(t.Asin)
	if err != nil {
		flog.Log.Errorf("AwsRunCategoryTask err:%s", err.Error())
		resp.Error = Error(SpiderBusyError, err.Error())
		return
	}

	resp.Flag = true
	return
}

type AwsSearchAsinRequest struct {
	KeyWord string `json:"key_word" validate:"required"`
}

func AwsSearchAsin(c *gin.Context) {
	return
}

func AwsMonitor() {
	return
}
