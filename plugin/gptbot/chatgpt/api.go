package chatgpt

import (
	"bytes"
	"encoding/json"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"gptbot/plugin/gptbot/constants"
	"gptbot/plugin/gptbot/model"
	"io"
	"net/http"
)

const (
	ApiPath = "/v1/chat/completions"
	Method  = constants.PostMethod
)

type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []*model.Message `json:"messages"`
}

type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Usage   Usage     `json:"usage"`
	Choices []Choices `json:"choices"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choices struct {
	Message      model.Message `json:"message"`
	FinishReason string        `json:"finish_reason"`
	Index        int           `json:"index"`
}

func (c *Client) fetchNextChatAnswer(req ChatRequest) (*model.Message, error) {
	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Transport: &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	request, err := http.NewRequest(Method, c.host+ApiPath, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", constants.DefaultContentType)
	request.Header.Add("Authorization", c.authorizationKey)

	var resp ChatResponse
	if err = retry.Do(func() error {
		r, err := client.Do(request)
		if err != nil {
			return err
		}
		defer r.Body.Close()
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return errors.Wrap(err, "Read body failed")
		}
		if r.StatusCode != 200 {
			return errors.New("Error status code" + string(bodyBytes))
		}
		if err = json.Unmarshal(bodyBytes, &resp); err != nil {
			return err
		}
		return nil
	},
		retry.Attempts(3),
	); err != nil {
		return nil, err
	}
	return &resp.Choices[0].Message, nil
}
