package chat

import (
	"crypto/rand"
	"encoding/binary"
	"github.com/pkg/errors"
	Constants "gptbot/plugin/gptbot/constants"
	"gptbot/plugin/gptbot/model"
	"math"
	"sync"
)

const (
	MaxContextCount = 10
)

type Chat struct {
	messages    []*model.Message
	Prompt      model.Message
	mutex       sync.RWMutex
	Model       string
	Probability int
}

func NewChat() *Chat {
	return &Chat{
		messages:    []*model.Message{},
		Prompt:      model.Message{Role: "system", Content: Constants.DefaultPrompt, Name: "system"},
		mutex:       sync.RWMutex{},
		Model:       Constants.GPT3DOT5MODEL,
		Probability: 0,
	}
}

func (c *Chat) SetPrompt(prompt string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Prompt = model.Message{
		Role:    "system",
		Content: prompt,
		Name:    "system",
	}
	c.messages = []*model.Message{}
}

func (c *Chat) GetPrompt() *model.Message {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return &c.Prompt
}

func (c *Chat) SetModel(model string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch model {
	case Constants.GPT3DOT5MODEL:
		c.Model = Constants.GPT3DOT5MODEL
	case Constants.GPT4MODEL:
		c.Model = Constants.GPT4MODEL
	case Constants.GPT4OMODEL:
		c.Model = Constants.GPT4OMODEL
	default:
		return errors.New("Unexpected gpt model:" + model)
	}
	return nil
}

func (c *Chat) GetModel() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Model
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
	return math.Abs(float64(n%100)) < float64(c.Probability)
}

func (c *Chat) ClearMessages() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messages = []*model.Message{}
}

func (c *Chat) SetGroupProbability(probability int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Probability = probability
}

func (c *Chat) GetGroupProbability() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Probability
}
