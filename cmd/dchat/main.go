package main

import (
	"github.com/dreamvids/dchat/pkg/chat"
	"log"
)

const (
	Name    string = "DreamVids Chat Server"
	Version string = "1.0"
)

func main() {
	log.Println("Hello world !")
	log.Println(Name, "- Version", Version)

	err := chat.ParseConfig("server.json")
	if err != nil {
		log.Fatalf("Fatal error while parsing config: %s", err)
	}

	err = chat.Start()
	if err != nil {
		log.Fatalf("Fatal error when running server: %s", err)
	}
}
