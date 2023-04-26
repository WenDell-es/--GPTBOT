package botService

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gptbot/src/config"
	"gptbot/src/goCQHttp"
	"gptbot/src/log"
	"gptbot/src/openAI/gpt"
	"gptbot/src/util"
	"io"
	"net/http"
	"strings"
)

type BotReq struct {
	MessageType string `json:"message_type"`
	SelfId      int64  `json:"self_id"`
	UserId      int64  `json:"user_id"`
	GroupId     int64  `json:"group_id"`
	Message     string `json:"message"`
}

type BotServer struct {
	Port         string
	Logger       *logrus.Logger
	GPTClient    *gpt.ChatGptClient
	CQHttpClient *goCQHttp.CQHttpClient
}

func (s *BotServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.Logger.Infoln(errors.Cause(err))
		return
	}
	req := BotReq{}
	if err = json.Unmarshal(body, &req); err != nil {
		s.Logger.Infoln(errors.Cause(err))
		return
	}
	if req.MessageType == "group" && !strings.HasPrefix(req.Message, util.GenerateAtCQCode(req.SelfId)) {
		return
	}
	req.Message = util.CutPrefixAndTrimSpace(req.Message, util.GenerateAtCQCode(req.SelfId))
	switch req.MessageType {
	case "private":
		s.handlePrivateMessage(req)
	case "group":
		s.handleGroupMessage(req)
	default:
		s.Logger.Errorln("Unsupported message type:", req.MessageType)
	}

}

func (s *BotServer) handlePrivateMessage(req BotReq) {
	resp := s.HandleOperation(req)
	if len(resp) > 0 {
		if _, err := s.CQHttpClient.SendPrivateMessage(req.UserId, resp); err != nil {
			s.Logger.Errorln(errors.Cause(err))
		}
		return
	}
	answer, err := s.GPTClient.QuestGpt(req.UserId, req.Message)
	if err != nil {
		s.Logger.Errorln(errors.Cause(err))
		answer = errors.Cause(err).Error()
	}
	if _, err := s.CQHttpClient.SendPrivateMessage(req.UserId, answer); err != nil {
		s.Logger.Errorln(errors.Cause(err))
	}
	s.Logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"user_id":      req.UserId,
		"group_id":     req.GroupId,
		"question":     req.Message,
		"answer":       answer,
		"prompt":       s.GPTClient.GetPrompt(req.UserId),
	}).Infoln()
}

func (s *BotServer) handleGroupMessage(req BotReq) {
	resp := s.HandleOperation(req)
	if len(resp) > 0 {
		if _, err := s.CQHttpClient.SendGroupMessage(req.GroupId, req.UserId, resp); err != nil {
			s.Logger.Errorln(errors.Cause(err))
		}
		return
	}
	answer, err := s.GPTClient.QuestGpt(req.GroupId, req.Message)

	if err != nil {
		s.Logger.Errorln(errors.Cause(err))
		answer = errors.Cause(err).Error()
	}
	if _, err := s.CQHttpClient.SendGroupMessage(req.GroupId, req.UserId, answer); err != nil {
		s.Logger.Errorln(errors.Cause(err))
	}
	s.Logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"user_id":      req.UserId,
		"group_id":     req.GroupId,
		"question":     req.Message,
		"answer":       answer,
		"prompt":       s.GPTClient.GetPrompt(req.GroupId),
	}).Infoln()
}

func NewBotServer(cfg config.Config) *BotServer {
	return &BotServer{
		Port:         cfg.Service.Port,
		Logger:       log.InitLog(),
		GPTClient:    gpt.NewChatGptClient(cfg.OpenAI, cfg.Proxy),
		CQHttpClient: goCQHttp.NewCQHttpClient(cfg.CQHttp),
	}
}

func (s *BotServer) Start() {
	http.HandleFunc("/", s.handleRequest)
	s.Logger.Infoln("GPT Bot Start")
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		panic(err)
	}
}
