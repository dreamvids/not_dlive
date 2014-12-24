package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	websocket.Conn

	Id        int
	Ready     bool
	Name      string
	Channel   string
	SessionId string
}

func (this *Client) Process() error {
	for {
		frame, err := this.ReceiveFrame()
		if err != nil {
			return &ProtocolError{err.Error()}
		}

		switch frame.Type {
		case websocket.TextMessage:
			err = HandleFrame(&frame, this)
			if err != nil {
				return err
			}
		case websocket.CloseMessage:
			this.Close()
		default:
			return InvalidMessageError
		}
	}

	return nil
}

func (this *Client) Initialize(name, channel, sessionId string) {
	this.Name = name
	this.Channel = channel
	this.SessionId = sessionId
	this.Ready = true
}

func (this *Client) ReceiveFrame() (Frame, error) {
	frameType, frameData, err := this.ReadMessage()
	if err != nil {
		return Frame{}, err
	}

	return Frame{frameType, frameData}, nil
}

func (this *Client) SendFrame(frame *Frame) error {
	return this.WriteMessage(frame.Type, frame.Data)
}

func (this *Client) SendMessage(message *Message) error {
	data, err := message.ToJson()
	if err != nil {
		return err
	}

	return this.SendFrame(&Frame{websocket.TextMessage, data})
}

func (this *Client) SendTextMessage(text string) error {
	message := Message{
		Sender:    "server",
		Type:      TextMessage,
		Content:   text,
		Timestamp: int64(time.Now().Unix()),
	}

	err := this.SendMessage(&message)
	if err != nil {
		return err
	}

	return nil
}

func (this *Client) GetRankStr() string {
	stmtUserId, err := Database.Prepare("SELECT user_id FROM users_sessions WHERE session_id=?")
	if err != nil {
		return ""
	}
	defer stmtUserId.Close()

	var userId int64
	err = stmtUserId.QueryRow(this.SessionId).Scan(&userId)
	if err != nil {
		return ""
	}

	stmtUser, err := Database.Prepare("SELECT username, rank FROM users WHERE id=?")
	if err != nil {
		return ""
	}
	defer stmtUser.Close()

	var username string
	var userRank int
	err = stmtUser.QueryRow(userId).Scan(&username, &userRank)
	if err != nil {
		return ""
	}

	switch userRank {
	case 4:
		return "modo"
	case 5:
		return "admin"
	default:
		return ""
	}

	stmtChannels, err := Database.Prepare("SELECT id FROM users_channels WHERE admins_ids LIKE ?")
	if err != nil {
		return ""
	}
	defer stmtChannels.Close()

	rows, err := stmtChannels.Query(fmt.Sprintf(";%d;", userId))
	if err != nil {
		return ""
	}

	var channelIds string
	for rows.Next() {
		var channelId string
		err = rows.Scan(&channelId)

		channelIds += fmt.Sprintf("%s; ", channelId)
	}

	channelIds = strings.TrimSuffix(channelIds, "; ")

	stmtLives, err := Database.Prepare("SELECT * FROM live_accesses WHERE channel_id IN (?)")
	if err != nil {
		return ""
	}
	defer stmtLives.Close()

	lives, err := stmtLives.Query(channelIds)
	if err != nil {
		return ""
	}

	for lives.Next() { // If a live acces was found, return true
		return "modo"
	}

	return ""
}

func (this *Client) isMuted() bool {
	stmtMuted, err := Database.Prepare("SELECT id FROM chat_mutes WHERE username = ?")
	defer stmtMuted.Close()

	muted, err := stmtMuted.Query(this.Name)
	if err != nil {
		return false
	}

	for muted.Next() {
		return true
	}

	return false
}

func (this *Client) GetUserId() (int64, error) {
	if this.SessionId == "" {
		return 0, fmt.Errorf("The client %s is not initialized !", this.Name)
	}

	stmtUserId, _ := Database.Prepare("SELECT user_id FROM users_sessions WHERE session_id=?")
	defer stmtUserId.Close()

	var userId int64
	err := stmtUserId.QueryRow(this.SessionId).Scan(&userId)
	if err != nil {
		return 0, err
	}

	return userId, nil
}
