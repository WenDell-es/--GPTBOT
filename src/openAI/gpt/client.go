package gpt

import (
	"github.com/pkg/errors"
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
		Model:    Constants.GPT3DOT5MODEL,
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
		Model:    chat.Model,
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
	chatLoader, ok := c.chats.Load(userId)
	if !ok {
		chatLoader = newChat()
	}
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

func (c *ChatGptClient) SetModel(userId int64, model string) error {
	chatLoader, ok := c.chats.Load(userId)
	if !ok {
		chatLoader = newChat()
	}
	chat := chatLoader.(*Chat)
	chat.Mutex.Lock()
	defer chat.Mutex.Unlock()
	switch model {
	case Constants.GPT3DOT5MODEL:
		chat.Model = Constants.GPT3DOT5MODEL
	case Constants.GPT4MODEL:
		chat.Model = Constants.GPT4MODEL
	default:
		return errors.New("Unexpected gpt model:" + model)
	}
	c.chats.Store(userId, chat)
	return nil
}

func (c *ChatGptClient) GetModel(userId int64) string {
	chatLoader, exist := c.chats.Load(userId)
	if !exist {
		chatLoader = newChat()
		return ""
	}
	chat := chatLoader.(*Chat)
	chat.Mutex.Lock()
	defer chat.Mutex.Unlock()
	return chat.Model
}
