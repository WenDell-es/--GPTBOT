package gpt

import "sync"

type Chat struct {
	Messages []*Message
	Prompt   Message
	Mutex    sync.RWMutex
	Model    string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
