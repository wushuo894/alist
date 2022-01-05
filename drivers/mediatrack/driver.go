package mediatrack

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

type MediaTrack struct{}

func (driver MediaTrack) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "MediaTrack",
	}
}

func (driver MediaTrack) Items() []base.Item {
	return []base.Item{
		{
			Name:        "access_token",
			Label:       "Token",
			Type:        base.TypeString,
			Description: "Unknown expiration time",
		},
		{
			Name:  "root_folder",
			Label: "root folder file_id",
			Type:  base.TypeString,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Values:   "updated_at,title,size",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "desc",
			Type:     base.TypeSelect,
			Values:   "true,false",
			Required: true,
		},
	}
}

func (driver MediaTrack) Save(account *model.Account, old *model.Account) error {
	return nil
}

func (driver MediaTrack) File(path string, account *model.Account) (*model.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &model.File{
			Id:        account.RootFolder,
			Name:      account.Name,
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: account.UpdatedAt,
		}, nil
	}
	dir, name := filepath.Split(path)
	files, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver MediaTrack) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
	} else {
		file, err := driver.File(path, account)
		if err != nil {
			return nil, err
		}
		files, err = driver.GetFiles(file.Id, account)
		if err != nil {
			return nil, err
		}
		if len(files) > 0 {
			_ = base.SetCache(path, files, account)
		}
	}
	return files, nil
}

func (driver MediaTrack) Link(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	pre := "https://jayce.api.mediatrack.cn/v3/assets/" + file.Id
	body, err := driver.Request(pre+"/token", base.Get, nil, nil, nil, nil, nil, account)
	if err != nil {
		return nil, err
	}
	url := pre + "/raw"
	res, err := base.NoRedirectClient.R().SetQueryParam("token", jsoniter.Get(body, "data").ToString()).Get(url)
	return &base.Link{Url: res.Header().Get("location")}, nil
}

func (driver MediaTrack) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("MediaTrack path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver MediaTrack) Proxy(c *gin.Context, account *model.Account) {

}

func (driver MediaTrack) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotImplement
}

func (driver MediaTrack) MakeDir(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://jayce.api.mediatrack.cn/v4/assets/%s/children", parentFile.Id)
	_, err = driver.Request(url, base.Post, nil, nil, nil, base.Json{
		"type":  1,
		"title": utils.Base(path),
	}, nil, account)
	return err
}

func (driver MediaTrack) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"parent_id": dstParentFile.Id,
		"ids":       []string{srcFile.Id},
	}
	url := "https://jayce.api.mediatrack.cn/v4/assets/batch/move"
	_, err = driver.Request(url, base.Post, nil, nil, nil, data, nil, account)
	return err
}

func (driver MediaTrack) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	url := "https://jayce.api.mediatrack.cn/v3/assets/" + srcFile.Id
	data := base.Json{
		"title": utils.Base(dst),
	}
	_, err = driver.Request(url, base.Put, nil, nil, nil, data, nil, account)
	return err
}

func (driver MediaTrack) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstParentFile, err := driver.File(utils.Dir(dst), account)
	if err != nil {
		return err
	}
	data := base.Json{
		"parent_id": dstParentFile.Id,
		"ids":       []string{srcFile.Id},
	}
	url := "https://jayce.api.mediatrack.cn/v4/assets/batch/clone"
	_, err = driver.Request(url, base.Post, nil, nil, nil, data, nil, account)
	return err
}

func (driver MediaTrack) Delete(path string, account *model.Account) error {
	parentFile, err := driver.File(utils.Dir(path), account)
	if err != nil {
		return err
	}
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	data := base.Json{
		"origin_id": parentFile.Id,
		"ids":       []string{file.Id},
	}
	url := "https://jayce.api.mediatrack.cn/v4/assets/batch/delete"
	_, err = driver.Request(url, base.Delete, nil, nil, nil, data, nil, account)
	return err
}

func (driver MediaTrack) Upload(file *model.FileStream, account *model.Account) error {
	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}
	src := "assets/56461b45-6d08-40a0-bfbf-0a47689bffaa" // random?
	var resp UploadResp
	_, err = driver.Request("https://jayce.api.mediatrack.cn/v3/storage/tokens/asset", base.Get, nil, map[string]string{
		"src": src,
	}, nil, nil, &resp, account)
	if err != nil {
		return err
	}
	credential := resp.Data.Credentials
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(credential.TmpSecretID, credential.TmpSecretKey, credential.Token),
		Region:      &resp.Data.Region,
		Endpoint:    aws.String("cos.accelerate.myqcloud.com"),
	}
	s, err := session.NewSession(cfg)
	if err != nil {
		return err
	}
	uploader := s3manager.NewUploader(s)
	input := &s3manager.UploadInput{
		Bucket: &resp.Data.Bucket,
		Key:    &resp.Data.Object,
		Body:   file,
	}
	_, err = uploader.Upload(input)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://jayce.api.mediatrack.cn/v3/assets/%s/children", parentFile.Id)
	data := base.Json{
		"category":    10,
		"description": file.GetFileName(),
		"hash":        "c459fda4d79554d692555932a5d5e6c1",
		"mime":        file.MIMEType,
		"size":        file.GetSize(),
		"src":         src,
		"title":       file.GetFileName(),
		"type":        0,
	}
	_, err = driver.Request(url, base.Post, nil, nil, nil, data, nil, account)
	return err
}

var _ base.Driver = (*MediaTrack)(nil)
