package chat

import (
	"encoding/json"
)

const (
	CommandMessage string = "command"
	TextMessage    string = "text"
)

// A frame represents a WebSocket packet
type Frame struct {
	Type int
	Data []byte
}

// Parse a WebSocket frame into a chat message
func (this *Frame) Parse() (Message, error) {
	var message Message

	if err := json.Unmarshal(this.Data, &message); err != nil {
		return message, InvalidMessageError
	}

	return message, nil
}

// A message sent by a chat client
type Message struct {
	Sender    string `json:"sender_name"`
	Type      string `json:"messageType"`
	Rank      string `json:"rank"`
	Channel   string `json:"channel"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	SessionId string `json:"sessionId"`
}

// Get the JSON representation of the message
func (this *Message) ToJson() ([]byte, error) {
	jsonData, err := json.Marshal(this)
	if err != nil {
		return nil, InvalidMessageError
	}

	return jsonData, nil
}
