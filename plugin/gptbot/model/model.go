package model

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name"`
}
