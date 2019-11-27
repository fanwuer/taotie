package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"io/ioutil"
	"math"
	"path/filepath"
	"taotie/core/config"
	. "taotie/core/flog"
	"taotie/core/model"
	myutil "taotie/core/util"
	"taotie/core/util/oss"
	"time"
)

var scaleType = []string{"jpg", "jpeg", "png"}
var FileAllow = map[string][]string{
	"image": {
		"jpg", "jpeg", "png", "gif"},
	"flash": {
		"swf", "flv"},
	"media": {
		"swf", "flv", "mp3", "wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb"},
	"file": {
		"doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2", "pdf"},
	"other": {
		"jpg", "jpeg", "png", "bmp", "gif", "swf", "flv", "mp3",
		"wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb",
		"doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2"}}

var (
	FileBytes = 1 << 25 // (1<<25)/1000.0/1000.0 33.54 size can not beyond 33M
)

type UploadResponse struct {
	FileName       string `json:"file_name"`
	ReallyFileName string `json:"really_file_name"`
	Size           int64  `json:"size"`
	Url            string `json:"url"`
	IsPicture      bool   `json:"is_picture"`
	Addon          string `json:"addon"`
	Oss            bool   `json:"oss"`
}

/*
	file: the binary file of HTML form's name
	type: can be: image、flash、media、file、other
	describe: some describe of file
*/
func UploadFile(c *gin.Context) {
	resp := new(Resp)
	data := UploadResponse{}
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	uu, err := GetUserSession(c)
	if err != nil {
		Log.Errorf("upload err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	uName := uu.Name

	fileType := c.DefaultPostForm("type", "other")
	if fileType == "" {
		fileType = "other"
	}

	tag := c.DefaultPostForm("tag", "other")
	if tag == "" {
		tag = "other"
	}

	describe := c.DefaultPostForm("describe", "")

	// Read binary file
	h, err := c.FormFile("file")
	if err != nil {
		Log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	fileAllowArray, ok := FileAllow[fileType]
	if !ok {
		Log.Errorf("upload err: type not permit")
		resp.Error = Error(UploadFileTypeNotPermit, "")
		return
	}

	fileSuffix := myutil.GetFileSuffix(h.Filename)
	if !myutil.InArray(fileAllowArray, fileSuffix) {
		Log.Errorf("upload err: file suffix: %s not permit", fileSuffix)
		resp.Error = Error(UploadFileTypeNotPermit, fmt.Sprintf("file suffix: %s not permit", fileSuffix))
		return
	}

	if h.Size > int64(FileBytes) {
		Log.Errorf("upload err: file size too big: %d", h.Size)
		resp.Error = Error(UploadFileTooMaxLimit, fmt.Sprintf(" file size too big: %d", h.Size))
		return
	}

	// Open file
	f, err := h.Open()
	if err != nil {
		Log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	defer f.Close()

	// Read binary
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		Log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	// When raw bytes empty will occur err
	fileSize := len(raw)
	if fileSize == 0 {
		Log.Errorf("upload err:%s", "file empty")
		resp.Error = Error(UploadFileError, "file empty")
		return
	}

	// HashCode the raw bytes
	fileHashCode, err := myutil.Sha256(raw)
	if err != nil {
		Log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	// HashCode add a prefix of userName, so diff user can upload the same file but the same user will still keep one file
	fileHashCode = uName + "_" + fileHashCode
	fileName := fileHashCode + "." + fileSuffix

	// If db exist the hash code of file
	p := new(model.File)
	p.HashCode = fileHashCode
	exist, err := p.Get()
	if err != nil {
		resp.Error = Error(DBError, err.Error())
		return
	}

	helpPath := fmt.Sprintf("storage/%s/%s", uName, fileType)
	if !exist {
		// File not exist, save it
		fileDir := filepath.Join(config.GlobalConfig.DefaultConfig.StoragePath, uName, fileType)
		fileAbName := filepath.Join(fileDir, fileName)

		// Local mode will save in disk
		if !config.GlobalConfig.DefaultConfig.IsOss {
			// disk mode first make dir
			err := myutil.MakeDir(fileDir)
			if err != nil {
				Log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}

			err = myutil.SaveToFile(fileAbName, raw)
			if err != nil {
				Log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}

			p.Url = fmt.Sprintf("/%s/%s", helpPath, fileName)
		} else {
			// Oss mode
			p.StoreType = 1
			p.Url = fmt.Sprintf("%s.%s/%s/%s", config.GlobalConfig.OssConfig.BucketName, config.GlobalConfig.OssConfig.Endpoint, helpPath, fileName)
			err = oss.SaveFile(config.GlobalConfig.OssConfig, helpPath+"/"+fileName, raw)
			if err != nil {
				Log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}
		}

		p.UrlHashCode, _ = myutil.Sha256([]byte(p.Url))

		// If is picture, cut the size
		if myutil.InArray(scaleType, fileSuffix) {
			p.IsPicture = 1
		}

		p.Type = fileType
		p.FileName = fileName
		p.ReallyFileName = h.Filename
		p.CreateTime = time.Now().Unix()
		p.Describe = describe
		p.UserId = uu.Id
		p.UserName = uName
		p.Tag = tag
		p.Size = int64(fileSize)
		_, err = model.Rdb.InsertOne(p)
		if err != nil {
			Log.Errorf("upload err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	} else {
		// File exist
		data.Addon = "file the same in server"
		if p.Status != 0 {
			// If file is hide must change back
			p.Status = 0
			p.UpdateStatus()
		}
	}

	// Return all basic info
	data.FileName = p.FileName
	data.ReallyFileName = p.ReallyFileName
	data.IsPicture = p.IsPicture == 1
	data.Size = p.Size
	data.Url = p.Url
	data.Oss = p.StoreType == 1
	resp.Data = data
	resp.Flag = true
	return
}

type ListFileAdminRequest struct {
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	SizeBegin       int64    `json:"size_begin"`
	SizeEnd         int64    `json:"size_end"`
	Sort            []string `json:"sort"`
	HashCode        string   `json:"hash_code"`
	Url             string   `json:"url"`
	StoreType       int      `json:"store_type" validate:"oneof=-1 0 1"`
	Status          int      `json:"status" validate:"oneof=-1 0 1"`
	Type            string   `json:"type"`
	Tag             string   `json:"tag"`
	UserId          int64    `json:"user_id"`
	Id              int64    `json:"id"`
	IsPicture       int      `json:"is_picture" validate:"oneof=-1 0 1"`
	PageHelp
}

type ListFileAdminResponse struct {
	Files []model.File `json:"files"`
	PageHelp
}

func ListFileAdminHelper(c *gin.Context, userId int64) {
	resp := new(Resp)

	respResult := new(ListFileAdminResponse)
	req := new(ListFileAdminRequest)
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
		Log.Errorf("ListFileAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.Rdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.File)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}
	if req.HashCode != "" {
		session.And("hash_code=?", req.HashCode)
	}

	if req.Status != -1 {
		// this not expose out, all people well see the file in show, those hide will emm, hide.
		session.And("status=?", req.Status)
	}

	if req.Url != "" {
		urlHashCode, _ := myutil.Sha256([]byte(req.Url))
		session.And("url_hash_code=?", urlHashCode)
	}

	if req.IsPicture != -1 {
		session.And("is_picture=?", req.IsPicture)
	}

	if req.Type != "" {
		session.And("type=?", req.Type)
	}

	if req.StoreType != -1 {
		session.And("store_type=?", req.StoreType)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	if userId != 0 {
		session.And("user_id=?", userId)
	} else {
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
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
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.SizeBegin > 0 {
		session.And("size>=?", req.SizeBegin)
	}

	if req.SizeEnd > 0 {
		session.And("size<?", req.SizeEnd)
	}

	// count num
	countSession := session.Clone()
	defer countSession.Close()
	total, err := countSession.Count()
	if err != nil {
		Log.Errorf("ListFileAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// if count>0 start list
	files := make([]model.File, 0)
	p := &req.PageHelp
	if total == 0 {
		if p.Limit == 0 {
			p.Limit = 20
		}
	} else {
		// sql build
		p.build(session, req.Sort, model.FileSortName)
		// do query
		err = session.Find(&files)
		if err != nil {
			Log.Errorf("ListFileAdmin err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
	}

	// result
	respResult.Files = files
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

// List all file info of all user, admin url
func ListFileAdmin(c *gin.Context) {
	ListFileAdminHelper(c, 0)
}

// list file of oneself
func ListFile(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		Log.Errorf("ListFile err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListFileAdminHelper(c, uid)
}

type UpdateFileRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Tag      string `json:"tag"`
	Hide     bool   `json:"hide"`
	Describe string `json:"describe"`
}

func UpdateFileAdminHelper(c *gin.Context, userId int64) {
	resp := new(Resp)
	req := new(UpdateFileRequest)
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
		Log.Errorf("UpdateFileAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	f := new(model.File)
	f.Id = req.Id

	// can change file tag so can group out
	f.Tag = req.Tag
	f.Describe = req.Describe
	f.UserId = userId

	// can set the file hide but it is pretend hide still exist
	ok, err := f.Update(req.Hide)
	if err != nil {
		Log.Errorf("UpdateFileAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	resp.Data = ok
	resp.Flag = true
}

// update file info or every user, admin url
func UpdateFileAdmin(c *gin.Context) {
	UpdateFileAdminHelper(c, 0)
}

// update file info of oneself
func UpdateFile(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		Log.Errorf("UpdateFile err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	UpdateFileAdminHelper(c, uid)
}
