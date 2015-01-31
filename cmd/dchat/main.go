package main

import (
	"flag"
	"github.com/dreamvids/chat/pkg/chat"
	"log"
)

const (
	Name    string = "DreamVids Chat Server"
	Version string = "1.0"
)

func main() {
	log.Println("Hello world !")
	log.Println(Name, "- Version", Version)

	configPath := flag.String("config", "dchat.json", "Configuration file to use")
	flag.Parse()

	err := chat.ParseConfig(*configPath)
	if err != nil {
		log.Fatalf("Fatal error while parsing config: %s", err)
	}

	err = chat.Start()
	if err != nil {
		log.Fatalf("Fatal error when running server: %s", err)
	}
}
