package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/dreamvids/dchat/pkg/chat"
)

const (
	Name    string = "DreamVids Chat Server"
	Version string = "1.0"
)

func main() {
	log.Println("Hello world !")
	log.Println(Name, "- Version", Version)

	port := flag.Int("port", 8080, "The port to listen to")
	flag.Parse()

	log.Println("Listening on port", *port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", chat.HandleWebsocket)
	http.Handle("/", mux)

	chat.ConnectToDatabase("127.0.0.1:3306", "root", "", "dreamvids_v2")
	defer chat.Database.Close()

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Cannot start ChatServer: ", err)
	}
}
