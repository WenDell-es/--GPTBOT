package gpt

import (
	"gptbot/src/config"
	Constants "gptbot/src/constants"
	"sync"
)

const (
	Bearer          = "Bearer "
	MaxContextCount = 10
)

type ChatGptClient struct {
	host             string
	chatAPIPath      string
	authorizationKey string
	chats            sync.Map
}

func newChat() *Chat {
	return &Chat{
		Messages: []*Message{},
		Prompt:   Message{"system", ""},
		Mutex:    sync.RWMutex{},
	}
}

func NewChatGptClient(openAiConfig config.OpenAIConfig, proxyConfig config.ProxyConfig) *ChatGptClient {
	chatGptClient := ChatGptClient{
		host:             openAiConfig.Host,
		chatAPIPath:      openAiConfig.ChatAPIPath,
		authorizationKey: Bearer + openAiConfig.AuthorizationKey,
		chats:            sync.Map{},
	}
	if proxyConfig.Enable {
		chatGptClient.host = proxyConfig.ProxyHost
	}
	return &chatGptClient
}

func (c *ChatGptClient) QuestGpt(userId int64, question string) (string, error) {
	chatLoader, exist := c.chats.Load(userId)
	if !exist {
		chatLoader = newChat()
	}
	chat := chatLoader.(*Chat)
	chat.Mutex.Lock()
	defer chat.Mutex.Unlock()
	for len(chat.Messages) >= MaxContextCount {
		chat.Messages = chat.Messages[2:]
	}
	chat.Messages = append(chat.Messages, &Message{
		Role:    "user",
		Content: question,
	})
	requestMessages := []*Message{&chat.Prompt}
	chatAnswer, err := c.fetchNextChatAnswer(ChatRequest{
		Model:    Constants.GPT3DOT5MODEL,
		Messages: append(requestMessages, chat.Messages...),
	})
	if err != nil {
		chat.Messages = chat.Messages[:len(chat.Messages)-1]
		return "", err
	}
	chat.Messages = append(chat.Messages, chatAnswer)
	c.chats.Store(userId, chat)

	return chatAnswer.Content, nil
}

func (c *ChatGptClient) SetPrompt(userId int64, prompt string) {
	chatLoader, _ := c.chats.Load(userId)
	chatLoader = newChat()
	chat := chatLoader.(*Chat)
	chat.Mutex.Lock()
	defer chat.Mutex.Unlock()
	chat.Prompt = Message{
		Role:    "system",
		Content: prompt,
	}
	c.chats.Store(userId, chat)
}

func (c *ChatGptClient) GetPrompt(userId int64) string {
	chatLoader, exist := c.chats.Load(userId)
	if !exist {
		chatLoader = newChat()
		return ""
	}
	chat := chatLoader.(*Chat)
	chat.Mutex.Lock()
	defer chat.Mutex.Unlock()
	return chat.Prompt.Content
}
