package chatgpt

import (
	"gptbot/plugin/gptbot/chat"
	"gptbot/plugin/gptbot/config"
	"gptbot/plugin/gptbot/model"
	"sync"
)

const (
	Bearer = "Bearer "
)

type Client struct {
	host             string
	chatAPIPath      string
	authorizationKey string
	chats            sync.Map
}

func NewChatGptClient(openAiConfig config.ChatGptConfig) *Client {
	chatGptClient := Client{
		host:             openAiConfig.Host,
		chatAPIPath:      openAiConfig.ChatAPIPath,
		authorizationKey: Bearer + openAiConfig.AuthorizationKey,
	}

	return &chatGptClient
}

func (c *Client) QuestGpt(currentChat *chat.Chat) (*model.Message, error) {
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
