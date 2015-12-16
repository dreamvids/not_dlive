package main

import (
	"log"
	"net/http"

	"github.com/dreamvids/dlive/chat"
	"github.com/dreamvids/dlive/stream"
	"github.com/gorilla/mux"
)

func handlePushStream(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	log.Println("Push", v["id"])

	stream.Push(v["id"], w, r)
}

func handlePullStream(w http.ResponseWriter, r *http.Request) {
	v := mux.Vars(r)
	log.Println("Pull", v["id"])

	stream.Push(v["id"], w, r)
}

func main() {
	log.Println("dlive server")

	r := mux.NewRouter()

	r.HandleFunc("/stream/push/{id}", handlePushStream)
	r.HandleFunc("/stream/pull/{id}", handlePullStream)
	http.Handle("/", r)

	err := chat.BindServer("/chat")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8000", nil))
}
