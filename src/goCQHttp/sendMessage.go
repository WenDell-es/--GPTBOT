package goCQHttp

import (
	"bytes"
	"encoding/json"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	Constants "gptbot/src/constants"
	"io"
	"net"
	"net/http"
)

const (
	APIPath = "/send_msg"
	Method  = Constants.PostMethod
)

type SendMessageReq struct {
	MessageType string `json:"message_type"`
	UserId      int64  `json:"user_id"`
	GroupId     int64  `json:"group_id"`
	Message     string `json:"message"`
	AutoEscape  bool   `json:"auto_escape"`
}
type SendMessageResp struct {
	MessageId int32 `json:"message_id"`
}

func (c *CQHttpClient) sendMessage(req SendMessageReq) (int32, error) {
	messageId := int32(-1)
	reqBody, err := json.Marshal(req)
	if err != nil {
		return messageId, errors.Wrap(err, "Json marshal failed")
	}
	request, err := http.NewRequest(Method, Constants.HttpPrefix+net.JoinHostPort(c.Host, c.Port)+APIPath, bytes.NewReader(reqBody))
	if err != nil {
		return messageId, errors.Wrap(err, "Create request failed")
	}
	request.Header.Add("Content-Type", Constants.DefaultContentType)
	sendMessageResp := SendMessageResp{}
	err = retry.Do(func() error {
		r, err := c.Client.Do(request)
		if err != nil {
			return errors.Wrap(err, "Do Http request failed")
		}
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Wrap(err, "Read body bytes failed")
		}
		if err = json.Unmarshal(bodyBytes, &sendMessageResp); err != nil {
			return errors.Wrap(err, "Json unmarshal sendMessageResp failed")
		}
		messageId = sendMessageResp.MessageId
		return nil
	},
		retry.Attempts(3),
	)
	return messageId, err
}
