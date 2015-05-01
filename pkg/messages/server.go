package messages

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	sio "github.com/googollee/go-socket.io"
	"github.com/jinzhu/gorm"
)

var (
	Database *gorm.DB    = nil
	Server   *sio.Server = nil

	MuteMessage string = "Muted by admin"
	BanMessage  string = "Banned by admin"

	ModoId  int = 4
	AdminId int = 5
)

func Init(db *sql.DB, muteMsg, banMsg string, modoId, adminId int) error {
	orm, err := gorm.Open("mysql", db)
	if err != nil {
		return err
	}
	Database = &orm

	s, err := sio.NewServer(nil)
	if err != nil {
		return err
	}
	Server = s

	MuteMessage = muteMsg
	BanMessage = banMsg
	ModoId = modoId
	AdminId = adminId

	Server.SetAllowRequest(func(r *http.Request) error { return nil })
	Server.On("connection", HandleConnection)
	Server.On("join", HandleJoin)
	Server.On("error", HandleError)

	return nil
}

func HandleConnection(s sio.Socket) {
	log.Println("Connection")

	s.On("chat message", func(msg string) {
		HandleChatMessage(s, msg)
	})
	s.On("command", func(msg string) {
		HandleCommand(s, msg)
	})
}

func HandleJoin(s sio.Socket, channel string) {
	s.Join(channel)
}

func HandleChatMessage(s sio.Socket, msg string) {
	if len(s.Rooms()) > 0 {
		m, err := ParseMessage(msg)
		if err != nil {
			log.Println(err)
			return
		}

		m.SessionId = ""

		data, err := m.ToJson()
		if err != nil {
			log.Println(err)
			return
		}

		Server.BroadcastTo(s.Rooms()[0], "chat message", string(data))
	}
}

func HandleCommand(s sio.Socket, command string) {
	msg, err := ParseMessage(command)
	if err != nil {
		return
	}

	args := strings.Split(msg.Content, " ")
	if len(args) <= 0 {
		args[0] = msg.Content
	}

	sessId := msg.SessionId

	switch args[0] {
	case "mute":
		if len(args) == 3 {
			processMute(sessId, args[1], args[2])
		}
		if len(args) == 2 {
			processMute(sessId, args[1], MuteMessage)
		}
		break
	case "ban":
		if len(args) == 3 {
			processBan(sessId, args[1], args[2])
		}
		if len(args) == 2 {
			processBan(sessId, args[1], BanMessage)
		}
		break
	case "rank":
		var str string
		rank := GetRank(sessId)

		switch rank {
		case RankModo:
			str = "Moderateur"
			break
		case RankAdmin:
			str = "Admin"
			break
		case RankStreamer:
			str = "Streamer"
			break
		case RankViewer:
			str = "Viewer"
			break
		}

		msg := MakeMessage("server", "Vous êtes un "+str)
		data, err := msg.ToJson()
		if err != nil {
			return
		}

		s.Emit("chat message", string(data))
		break
	}
}

func HandleError(s sio.Socket, err error) {
	log.Println("Error", err)
}

func processMute(sessId, username, reason string) {

}

func processBan(sessId, username, reason string) {

}
