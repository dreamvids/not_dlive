package messages

import (
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	Sender    string `json:"sender"`
	Rank      int    `json:"rank"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	SessionId string `json:"session_id"`
}

func ParseMessage(data string) (Message, error) {
	var msg Message

	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func MakeMessage(sender, content string) Message {
	msg := Message{
		Sender:    sender,
		Rank:      0,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	return msg
}

func (this *Message) ToJson() ([]byte, error) {
	jsonData, err := json.Marshal(this)
	if err != nil {
		return nil, fmt.Errorf("Invalid message")
	}

	return jsonData, nil
}
