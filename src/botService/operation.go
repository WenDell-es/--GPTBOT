package botService

import (
	"gptbot/src/util"
	"strings"
)

var operationList = []string{
	"-setPrompt",
	"-showPrompt",
	"-help",
	"-setGPTModel",
}

func (s *BotServer) HandleOperation(req BotReq) string {
	operationMap := map[string]func(req BotReq) string{
		"-setPrompt":   s.setPrompt,
		"-showPrompt":  s.getPrompt,
		"-help":        s.getHelpMessage,
		"-setGPTModel": s.setGPTModel,
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
	prompt := util.CutPrefixAndTrimSpace(req.Message, "-setPrompt")
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
	return "-help 显示帮助\n\n设置当前会话场景：\n-setPrompt 设置提示词\n-showPrompt 显示提示词\n-setGPTModel 设置gpt模型(gpt-3.5-turbo gpt-4)"
}

func (s *BotServer) setGPTModel(req BotReq) string {
	userId := req.GroupId
	if req.MessageType == "private" {
		userId = req.UserId
	}
	model := util.CutPrefixAndTrimSpace(req.Message, "-setGPTModel")
	if err := s.GPTClient.SetModel(userId, model); err != nil {
		return err.Error()
	}
	return "设置GPT模型成功"
}
