package chat

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/pkg/errors"
	Constants "gptbot/src/constants"
	"gptbot/src/model"
	"math"
	"sync"
)

const (
	MaxContextCount = 10
)

type Chat struct {
	messages    []*model.Message
	prompt      model.Message
	mutex       sync.RWMutex
	model       string
	probability int
}

func NewChat() *Chat {
	return &Chat{
		messages:    []*model.Message{},
		prompt:      model.Message{Role: "system", Content: Constants.DefaultPrompt, Name: "system"},
		mutex:       sync.RWMutex{},
		model:       Constants.GPT3DOT5MODEL,
		probability: 0,
	}
}

func (c *Chat) SetPrompt(prompt string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.prompt = model.Message{
		Role:    "system",
		Content: prompt,
		Name:    "system",
	}
	c.messages = []*model.Message{}
}

func (c *Chat) GetPrompt() *model.Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return &c.prompt
}

func (c *Chat) SetModel(model string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch model {
	case Constants.GPT3DOT5MODEL:
		c.model = Constants.GPT3DOT5MODEL
	case Constants.GPT4MODEL:
		c.model = Constants.GPT4MODEL
	default:
		return errors.New("Unexpected gpt model:" + model)
	}
	return nil
}

func (c *Chat) GetModel() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.model
}

func (c *Chat) AddMessage(message *model.Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for len(c.messages) > MaxContextCount {
		c.messages = c.messages[1:]
	}
	c.messages = append(c.messages, message)
}

func (c *Chat) RemoveLastMessage() {
	c.messages = c.messages[:len(c.messages)-1]
}

func (c *Chat) GetMessages() []*model.Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.messages
}

func (c *Chat) GroupChatCheck() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var n int32
	_ = binary.Read(rand.Reader, binary.LittleEndian, &n)
	math.Abs(float64(n % 100))
	return math.Abs(float64(n%100)) < float64(c.probability)
}

func (c *Chat) ClearMessages() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messages = []*model.Message{}
}

func (c *Chat) SetGroupProbability(probability int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.probability = probability
}

func (c *Chat) GetGroupProbability() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.probability
}
