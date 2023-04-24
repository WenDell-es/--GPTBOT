package botService

import (
	"gptbot/src/util"
	"strings"
)

var operationList = []string{
	"-set",
	"-show",
	"-help",
}

func (s *BotServer) HandleOperation(req BotReq) string {
	operationMap := map[string]func(req BotReq) string{
		"-set":  s.setPrompt,
		"-show": s.getPrompt,
		"-help": s.getHelpMessage,
	}
	for _, operation := range operationList {
		if strings.HasPrefix(req.Message, operation) {
			return operationMap[operation](req)
		}
	}
	return ""
}

func (s *BotServer) setPrompt(req BotReq) string {
	userId := req.GroupId
	if req.MessageType == "private" {
		userId = req.UserId
	}
	prompt := util.CutPrefixAndTrimSpace(req.Message, "-set")
	s.GPTClient.SetPrompt(userId, prompt)
	return "设置提示词成功:" + prompt
}

func (s *BotServer) getPrompt(req BotReq) string {
	userId := req.GroupId
	if req.MessageType == "private" {
		userId = req.UserId
	}
	return "当前提示词为:" + s.GPTClient.GetPrompt(userId)
}

func (s *BotServer) getHelpMessage(req BotReq) string {
	return "-help 显示帮助\n\n设置当前会话场景：\n-set 设置提示词\n-show 显示提示词"
}
