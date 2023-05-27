package botservice

import (
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"gptbot/plugin/gptbot/chat"
	"gptbot/plugin/gptbot/chatgpt"
	"gptbot/plugin/gptbot/config"
	"gptbot/plugin/gptbot/model"
	"gptbot/plugin/gptbot/util"
	"strconv"
	"sync"
)

type GptBot struct {
	GPTClient *chatgpt.Client
	Chats     sync.Map
}

func NewGptBot(cfg config.ChatGptConfig) *GptBot {
	return &GptBot{
		GPTClient: chatgpt.NewChatGptClient(cfg),
		Chats:     sync.Map{},
	}
}

func (b *GptBot) GetChat(id int64) *chat.Chat {
	chatLoader, exist := b.Chats.Load(id)
	if !exist {
		chatLoader = chat.NewChat()
		b.Chats.Store(id, chatLoader)
	}
	return chatLoader.(*chat.Chat)
}

func (b *GptBot) Talk(ctx *zero.Ctx) string {
	id := util.GetChatId(ctx)
	currentChat := b.GetChat(id)
	answer, err := b.GPTClient.QuestGpt(currentChat)
	// 无报错直接返回
	if answer != nil && err == nil {
		answer.Name = strconv.FormatInt(ctx.Event.SelfID, 10)
		currentChat.AddMessage(answer)
		return answer.Content
	}
	logrus.WithFields(logrus.Fields{"Event": ctx.Event, "History": currentChat.GetMessages(), "Prompt": currentChat.GetPrompt().Content, "model": currentChat.GetModel()}).Warnln("gpt api报错, 准备清除记忆并重试", err)
	// 报错则仅加载最近的一条聊天记录再次尝试
	currentChat.ClearMessages()
	currentChat.AddMessage(&model.Message{
		Role:    "user",
		Content: ctx.Event.Message.String(),
		Name:    strconv.FormatInt(ctx.Event.UserID, 10),
	})
	answer, err = b.GPTClient.QuestGpt(currentChat)
	// 依然出错则清空记忆区
	if answer == nil || err != nil {
		currentChat.ClearMessages()
		logrus.WithFields(logrus.Fields{"Event": ctx.Event, "History": currentChat.GetMessages(), "Prompt": currentChat.GetPrompt().Content, "model": currentChat.GetModel()}).Errorln("gpt api报错", err)
		return "你说的话太刺激了，猫猫被吓晕过去了。所以刚刚说到哪了？"
	}
	return answer.Content
}
