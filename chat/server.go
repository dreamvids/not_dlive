package chat

import (
	"github.com/googollee/go-socket.io"
	"net/http"
)

func BindServer(url string) error {
	cs, err := socketio.NewServer(nil)
	if err != nil {
		return err
	}

	cs.On("connection", func(s socketio.Socket) {
		s.Join("global")
		s.On("global message", func(msg string) {
		})
		s.On("disconnection", func() {
		})
	})
	cs.On("error", func(s socketio.Socket, err error) {
	})

	http.Handle(url, cs)
	return nil
}
