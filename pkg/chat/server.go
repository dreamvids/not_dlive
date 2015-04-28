package chat

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	Port       int      = 8080
	MaxClients int      = 0
	Clients    []Client = make([]Client, 0)
	Database   *sql.DB  = nil

	ModoRankId  int    = 4
	AdminRankId int    = 5
	MuteMessage string = "Vous n'avez pas été sage. Vous ne pouvez donc pas parler."

	nextClientId int = 1
)

func Start(db *sql.DB) error {
	Database = db

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleWebsocket)
	http.Handle("/", mux)

	log.Println("Listening on port", Port)
	addr := fmt.Sprintf("0.0.0.0:%d", Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		return fmt.Errorf("Cannot bind http server: %s", err)
	}

	return nil
}

func AddClient(addr string, conn *websocket.Conn) (*Client, error) {
	if MaxClients != 0 && len(Clients) >= MaxClients {
		return nil, &ServerError{fmt.Sprintf("Can not accept client [%s] : server full (max clients: %d)", conn.RemoteAddr(), MaxClients)}
	}

	var client Client = Client{*conn, nextClientId, false, "", "", ""}

	Clients = append(Clients, client)
	log.Println("Client", addr, "connected with id", nextClientId)

	nextClientId++
	return &Clients[len(Clients)-1], nil
}

func HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,

		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Client connection failed: ", err)
		return
	}

	client, err := AddClient(r.RemoteAddr, conn)
	if err != nil {
		log.Println(err)
	}

	err = client.Process()
	if err != nil {
		log.Println("Failed to process client: ", err)
	}

	log.Println("Client", client.Id, "lost connection")
	defer client.Close()
}

func HandleFrame(frame *Frame, client *Client) error {
	log.Println("Received frame: ", string(frame.Data))

	message, err := frame.Parse()
	if err != nil {
		return err
	}

	if !client.Ready {
		client.Initialize(message.Sender, message.Channel, message.SessionId)
		return nil
	}

	if client.SessionId == "" {
		client.SessionId = message.SessionId
	}

	if message.Type == TextMessage && message.Content != "" {
		if client.isMuted() {
			client.SendTextMessage(MuteMessage)
			return nil
		}

		var newMessage = Message{
			Sender:    message.Sender,
			Type:      TextMessage,
			Rank:      client.GetRankStr(),
			Channel:   message.Channel,
			Content:   message.Content,
			Timestamp: message.Timestamp,
		}

		jsonData, err := newMessage.ToJson()
		if err != nil {
			return err
		}

		BroadcastFrame(&Frame{websocket.TextMessage, jsonData}, message.Channel)
	}

	if message.Type == CommandMessage && message.Content != "" {
		ProcessCommand(&message, client)
	}

	return nil
}

func ProcessCommand(command *Message, client *Client) {
	rank := client.GetRankStr()
	if rank == "modo" || rank == "admin" {
		cmdStr := strings.TrimPrefix(command.Content, "/")
		args := strings.Split(cmdStr, " ")

		if len(args) < 1 {
			return
		}

		switch args[0] {
		case "kick":
			if len(args) == 3 {
				username := args[1]
				reason := args[2]

				err := processKick(username, reason, client)
				if err != nil {
					log.Printf("Failed processing kick command from %s: %s", client.Name, err)
					return
				}
			} else {
				errMsg := "Kick command usage: /kick <username> <reason>"
				client.SendTextMessage(errMsg)
			}
			break
		case "mute":
			if len(args) == 3 {
				username := args[1]
				reason := args[2]

				err := processMute(username, reason, client)
				if err != nil {
					log.Printf("Failed processing mute command from %s: %s", client.Name, err)
					return
				}
			} else {
				errMsg := "Mute command usage: /mute <username> <reason>"
				client.SendTextMessage(errMsg)
			}
			break
		case "unmute":
			if len(args) == 2 {
				username := args[1]

				err := processUnmute(username, client)
				if err != nil {
					log.Printf("Failed processing unmute command from %s: %s", client.Name, err)
					return
				}
			} else {
				errMsg := "Unmute command usage: /unmute <username>"
				client.SendTextMessage(errMsg)
			}
			break
		default:
			return
		}
	}
}

func processKick(username, reason string, admin *Client) error {
	client, err := GetClientByName(username)
	if err != nil {
		return err
	}

	client.SendTextMessage("You were kicked from the chat: " + reason)
	admin.SendTextMessage("Kicked " + username + ": " + reason)

	err = RemoveClient(client)
	if err != nil {
		return err
	}

	return nil
}

func processMute(username, reason string, admin *Client) error {
	client, err := GetClientByName(username)
	if err != nil {
		return err
	}

	insStmt, _ := Database.Prepare("INSERT INTO chat_mutes VALUES ('', ?, ?, ?)")
	defer insStmt.Close()

	adminId, err := admin.GetUserId()
	if err != nil {
		return &ServerError{"Can not authenticate admin " + admin.Name}
	}

	_, err = insStmt.Query(username, adminId, time.Now().Unix())

	err = client.SendTextMessage("You were muted by admin: " + reason)
	if err != nil {
		log.Println(err)
	}
	admin.SendTextMessage("Muted " + username + ": " + reason)

	return nil
}

func processUnmute(username string, admin *Client) error {
	delStmt, _ := Database.Prepare("DELETE FROM chat_mutes WHERE username = ?")
	defer delStmt.Close()

	_, err := admin.GetUserId()
	if err != nil {
		return &ServerError{"Can not authenticate admin " + admin.Name}
	}

	_, err = delStmt.Query(username)
	if err != nil {
		return &DatabaseError{err.Error()}
	}

	admin.SendTextMessage("Unmuted " + username)
	if err != nil {
		return err
	}

	client, err := GetClientByName(username)
	if err != nil {
		return nil
	}

	client.SendTextMessage("You were unmuted")
	return nil
}

func BroadcastFrame(frame *Frame, channel string) {
	for i := range Clients {
		if channel == Clients[i].Channel {
			Clients[i].SendFrame(frame)
		}
	}
}

func BroadcastFrameExcept(frame *Frame, channel string, excepted *Client) {
	for i := range Clients {
		client := Clients[i]

		if client.Id != excepted.Id && client.Channel == channel {
			client.SendFrame(frame)
		}
	}
}

func ClientNameAvailable(name string) bool {
	for i := range Clients {
		if Clients[i].Name == name {
			return false
		}
	}

	return true
}

func GetClientByName(name string) (*Client, error) {
	for i := range Clients {
		if Clients[i].Name == name {
			return &Clients[i], nil
		}
	}

	return &Client{}, &ServerError{fmt.Sprintf("Cannot find a client with name '%s'", name)}
}

func RemoveClient(client *Client) error {
	i := client.Id - 1

	if i >= len(Clients) {
		return &ServerError{"Can not find the specified client in client list"}
	}

	client.Close()
	Clients = append(Clients[:i], Clients[i+1:]...) // Remove the client from the list

	return nil
}
