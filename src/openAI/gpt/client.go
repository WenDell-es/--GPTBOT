package gpt

import (
	"gptbot/src/chat"
	"gptbot/src/config"
	"gptbot/src/model"
	"sync"
)

const (
	Bearer = "Bearer "
)

type ChatGptClient struct {
	host             string
	chatAPIPath      string
	authorizationKey string
	chats            sync.Map
}

func NewChatGptClient(openAiConfig config.OpenAIConfig, proxyConfig config.ProxyConfig) *ChatGptClient {
	chatGptClient := ChatGptClient{
		host:             openAiConfig.Host,
		chatAPIPath:      openAiConfig.ChatAPIPath,
		authorizationKey: Bearer + openAiConfig.AuthorizationKey,
	}
	if proxyConfig.Enable {
		chatGptClient.host = proxyConfig.ProxyHost
	}
	return &chatGptClient
}

func (c *ChatGptClient) QuestGpt(currentChat *chat.Chat) (*model.Message, error) {
	promptMessage := []*model.Message{currentChat.GetPrompt()}

	chatAnswer, err := c.fetchNextChatAnswer(ChatRequest{
		Model:    currentChat.GetModel(),
		Messages: append(promptMessage, currentChat.GetMessages()...),
	})
	if err != nil {
		return nil, err
	}
	return chatAnswer, nil
}
