package store

import (
	"context"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	*cos.Client
}

type Cfg struct {
	Host      string
	SecretID  string
	SecretKey string
}

const (
	SectionName = "cos"
	DefaultPath = "./config/config.ini"
	MaxKeys     = 1000
	Delimiter   = "/"
)

var storeClient *Client

func init() {
	sc, err := newStoreClient(DefaultPath)
	if err != nil {
		logrus.Fatalln("cos store初始化失败", err)
	}
	storeClient = sc
}

func GetStoreClient() *Client {
	return storeClient
}

func newStoreClient(path string) (*Client, error) {
	conf, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	cosCfg := &Cfg{}
	err = conf.Section(SectionName).MapTo(cosCfg)
	if err != nil {
		return nil, err
	}
	u, _ := url.Parse(cosCfg.Host)

	return &Client{cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosCfg.SecretID,
			SecretKey: cosCfg.SecretKey,
		},
	})}, nil
}

func (s *Client) FetchAllFileInfo(prefix string) ([]cos.Object, error) {
	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:    prefix,
		Delimiter: Delimiter,
		MaxKeys:   MaxKeys,
	}
	res := []cos.Object{}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := s.Bucket.Get(context.Background(), opt)
		if err != nil {
			return nil, err
		}
		res = append(res, v.Contents...)
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}
	// 去掉prefix/
	if len(res) > 0 {
		res = res[1:]
	}
	return res, nil
}

func (s *Client) GetObjectUrl(key string) string {
	return s.Object.GetObjectURL(key).String()
}

func (s *Client) GetObjectBytes(key string) ([]byte, error) {
	resp, err := s.Object.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}
	bs, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return bs, err
}

func (s *Client) UploadObject(filePath, name string) error {
	_, err := s.Object.PutFromFile(context.Background(), name, filePath, nil)
	return err
}

func (s *Client) DeleteObject(key string) error {
	_, err := s.Object.Delete(context.Background(), key, nil)
	return err
}

func (s *Client) IsExist(key string) (bool, error) {
	return s.Object.IsExist(context.Background(), key)
}
