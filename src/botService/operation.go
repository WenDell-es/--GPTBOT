package botService

import (
	"github.com/pkg/errors"
	"gptbot/src/chat"
	"gptbot/src/model"
	"gptbot/src/util"
	"strconv"
	"strings"
)

var operationList = []string{
	"-setPrompt",
	"-showPrompt",
	"-help",
	"-setGPTModel",
	"-getGPTModel",
	"-clearMessages",
	"-setGroupChatProbability",
	"-getGroupChatProbability",
	"-getMessages",
}

type UserMessage struct {
	chatId      int64
	userId      int64
	message     string
	messageType string
	selfId      int64
}

// @description   解析用户发来的请求，并做对应处理
// @param userMessage 用户发来的请求信息
// @param currentChat 当前用户对应的chat对象指针
// @return 返回机器人给用户的回复，若不需要回复则返回空字符串

func (s *BotServer) HandleOperation(userMessage UserMessage, currentChat *chat.Chat) string {
	var operationMap = map[string]func(UserMessage, *chat.Chat) string{
		"-setPrompt":               s.setPrompt,
		"-showPrompt":              s.getPrompt,
		"-help":                    s.getHelpMessage,
		"-setGPTModel":             s.setGPTModel,
		"-getGPTModel":             s.getGPTModel,
		"-clearMessages":           s.clearMessages,
		"-setGroupChatProbability": s.setGroupChatProbability,
		"-getGroupChatProbability": s.getGroupChatProbability,
		"-getMessages":             s.getMessages,
	}
	// 若信息中存在上面的指令前缀，则执行对应方法
	for _, operation := range operationList {
		if strings.HasPrefix(userMessage.message, operation) {
			// 去除message中的前缀和无用空白
			userMessage.message = util.CutPrefixAndTrimSpace(userMessage.message, operation)
			return operationMap[operation](userMessage, currentChat)
		}
	}
	// 若未在循环中退出，则说明不符合上述任何指令，为普通聊天内容
	return s.questGPT(userMessage, currentChat)
}

func (s *BotServer) setPrompt(userMessage UserMessage, chat *chat.Chat) string {
	chat.SetPrompt(userMessage.message)
	return "设置提示词成功:\n" + chat.GetPrompt().Content
}

func (s *BotServer) getPrompt(userMessage UserMessage, chat *chat.Chat) string {
	return "当前提示词为:\n" + chat.GetPrompt().Content
}

func (s *BotServer) getHelpMessage(userMessage UserMessage, chat *chat.Chat) string {
	return "-help 显示帮助\n\n设置当前会话场景：\n-setPrompt 设置提示词\n-showPrompt 显示提示词\n-setGPTModel 设置gpt模型(gpt-3.5-turbo gpt-4)\n-getGPTModel 查看当前gpt model" +
		"\n-clearMessages 清除聊天记录" +
		"\n-setGroupChatProbability 设置群聊回复频率，默认0，设置为非0数值后，机器人会对没有at机器人的聊天做出回应。0为不回复，100为回复每一条聊天" +
		"\n-getGroupChatProbability 查看群聊回复频率" +
		"\n-getMessages 查看当前记忆区"
}

func (s *BotServer) setGPTModel(userMessage UserMessage, chat *chat.Chat) string {
	if err := chat.SetModel(userMessage.message); err != nil {
		return err.Error()
	}
	return "设置GPT模型成功:\n" + chat.GetModel()
}

func (s *BotServer) getGPTModel(userMessage UserMessage, chat *chat.Chat) string {
	return "设置GPT模型成功:\n" + chat.GetModel()
}

func (s *BotServer) clearMessages(userMessage UserMessage, chat *chat.Chat) string {
	chat.ClearMessages()
	return "已清除聊天记忆区"
}

func (s *BotServer) getMessages(userMessage UserMessage, chat *chat.Chat) string {
	resp := ""
	messages := chat.GetMessages()
	for _, message := range messages {
		resp += message.Name + ":" + message.Content + "\n"
	}
	return resp
}

func (s *BotServer) setGroupChatProbability(userMessage UserMessage, chat *chat.Chat) string {
	probability, err := strconv.Atoi(userMessage.message)
	if err != nil {
		return err.Error()
	}
	chat.SetGroupProbability(probability)
	return "设置群聊概率为:\n" + strconv.Itoa(chat.GetGroupProbability())
}

func (s *BotServer) getGroupChatProbability(userMessage UserMessage, chat *chat.Chat) string {
	return "当前群聊概率为:\n" + strconv.Itoa(chat.GetGroupProbability())
}

func (s *BotServer) questGPT(userMessage UserMessage, currentChat *chat.Chat) string {
	// 删除所有CQ码，并去除无用空格
	content := strings.TrimSpace(util.RemoveAllCQCode(userMessage.message))
	if content == "" {
		return "" // 空内容直接返回
	}
	// 将message放入当前聊天记录中
	currentChat.AddMessage(&model.Message{
		Role:    "user",
		Content: content,
		Name:    strconv.FormatInt(userMessage.chatId, 10),
	})

	// 若满足以下条件：1.群聊。2.没有at机器人。3.随机条件判断失败 则仅记住本次聊天内容，不实际请求open ai
	if userMessage.messageType == "group" && !util.IsStringAboutMe(userMessage.message, userMessage.selfId) && !currentChat.GroupChatCheck() {
		return ""
	}
	prefix := ""
	// 若在群聊中at机器人，则机器人的回复也加上at用户的CR代码
	if userMessage.messageType == "group" && util.IsStringAboutMe(userMessage.message, userMessage.selfId) {
		prefix = util.GenerateAtCQCode(userMessage.userId)
	}

	// 请求gpt
	answer, err := s.GPTClient.QuestGpt(currentChat)
	if err != nil {
		s.Logger.Errorln(errors.Cause(err))
		// 请求错误则将失败的问题移出聊天记录
		currentChat.RemoveLastMessage()
		return err.Error()
	}
	// 将回答放入当前聊天记录中
	if answer != nil {
		answer.Name = strconv.FormatInt(userMessage.selfId, 10)
		currentChat.AddMessage(answer)
	}
	return prefix + answer.Content
}
