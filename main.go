package main

import (
	"log"
	"net/http"

	"github.com/dreamvids/dlive/chat"
	"github.com/dreamvids/dlive/stream"
	"github.com/gorilla/mux"
)

func handlePushStream(w http.ResponseWriter, r *http.Request) {
}

func handlePullStream(w http.ResponseWriter, r *http.Request) {
}

func main() {
	log.Println("dlive server")

	r := mux.NewRouter()

	notif := make(chan int)
	r.HandleFunc("/stream/push/{id}", func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]

		stream.Push(id, notif, w, r)
	})
	r.HandleFunc("/stream/pull/{id}", func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		id := v["id"]

		stream.Pull(id, notif, w, r)
	})

	http.Handle("/", r)

	err := chat.BindServer("/chat")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8000", nil))
}
