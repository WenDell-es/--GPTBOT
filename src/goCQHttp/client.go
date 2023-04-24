package goCQHttp

import (
	"crypto/tls"
	"gptbot/src/config"
	"gptbot/src/util"
	"net/http"
)

type CQHttpClient struct {
	Host   string
	Port   string
	Client http.Client
}

func NewCQHttpClient(cfg config.CQHttpConfig) *CQHttpClient {
	return &CQHttpClient{
		Host: cfg.Host,
		Port: cfg.Port,
		Client: http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}},
	}
}

func (c *CQHttpClient) SendPrivateMessage(userId int64, message string) (int32, error) {
	messageId, err := c.sendMessage(SendMessageReq{
		MessageType: "private",
		UserId:      userId,
		Message:     message,
		AutoEscape:  false,
	})
	return messageId, err
}

func (c *CQHttpClient) SendGroupMessage(groupId, userId int64, message string) (int32, error) {
	messageId, err := c.sendMessage(SendMessageReq{
		MessageType: "group",
		GroupId:     groupId,
		Message:     util.GenerateAtCQCode(userId) + message,
		AutoEscape:  false,
	})
	return messageId, err
}
