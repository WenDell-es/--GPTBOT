package botService

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gptbot/src/chat"
	"gptbot/src/config"
	Constants "gptbot/src/constants"
	"gptbot/src/goCQHttp"
	"gptbot/src/log"
	"gptbot/src/openAI/gpt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	Chats        sync.Map
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
	s.Logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"user_id":      req.UserId,
		"group_id":     req.GroupId,
		"question":     req.Message,
	}).Infoln()

	switch req.MessageType {
	case "private":
		s.handlePrivateMessage(req)
	case "group":
		s.handleGroupMessage(req)
	default:
		s.Logger.Errorln("Unsupported message type:", req.MessageType)
	}
	w.WriteHeader(200)
}

func (s *BotServer) handlePrivateMessage(req BotReq) {
	chatLoader, exist := s.Chats.Load(req.UserId)
	if !exist {
		chatLoader = chat.NewChat()
	}
	currentChat := chatLoader.(*chat.Chat)
	s.Chats.Store(req.UserId, currentChat)
	userMessage := UserMessage{req.UserId, req.UserId, req.Message, req.MessageType, req.SelfId}
	resp := s.HandleOperation(userMessage, currentChat)
	if len(resp) <= 0 {
		return
	}
	for _, str := range Constants.UnExpectedResp {
		if strings.Contains(resp, str) {
			resp = "换一个话题吧。。。"
			break
		}
	}
	if _, err := s.CQHttpClient.SendPrivateMessage(req.UserId, resp); err != nil {
		s.Logger.Errorln(errors.Cause(err))
	}
	s.Logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"user_id":      req.UserId,
		"group_id":     req.GroupId,
		"question":     req.Message,
		"answer":       resp,
		"prompt":       currentChat.GetPrompt(),
		"model":        currentChat.GetModel(),
	}).Infoln()

}

func (s *BotServer) handleGroupMessage(req BotReq) {
	chatLoader, exist := s.Chats.Load(req.GroupId)
	if !exist {
		chatLoader = chat.NewChat()
	}
	currentChat := chatLoader.(*chat.Chat)
	s.Chats.Store(req.GroupId, currentChat)
	userMessage := UserMessage{req.GroupId, req.UserId, req.Message, req.MessageType, req.SelfId}

	resp := s.HandleOperation(userMessage, currentChat)
	if len(resp) <= 0 {
		return
	}
	for _, str := range Constants.UnExpectedResp {
		if strings.Contains(resp, str) {
			resp = "换一个话题吧。。。"
			break
		}
	}
	if _, err := s.CQHttpClient.SendGroupMessage(req.GroupId, resp); err != nil {
		s.Logger.Errorln(errors.Cause(err))
	}
	s.Logger.WithFields(logrus.Fields{
		"message_type": req.MessageType,
		"user_id":      req.UserId,
		"group_id":     req.GroupId,
		"question":     req.Message,
		"answer":       resp,
		"prompt":       currentChat.GetPrompt(),
		"model":        currentChat.GetModel(),
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
