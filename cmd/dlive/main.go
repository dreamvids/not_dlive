package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dreamvids/dlive/pkg/chat"
	"github.com/dreamvids/dlive/pkg/events"
)

const (
	Name    string = "DreamVids Live Server"
	Version string = "1.0"
)

type serverConfig struct {
	Port        int
	MaxClients  int    `json:"chat-max-clients"`
	ModoRank    int    `json:"chat-modo-rank"`
	AdminRank   int    `json:"chat-admin-rank"`
	MuteMessage string `json:"chat-mute-message"`

	DbHost string `json:"db-host"`
	DbUser string `json:"db-username"`
	DbPass string `json:"db-password"`
	DbName string `json:"db-name"`
}

func parseConfig(path string) (serverConfig, error) {
	var config serverConfig

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func main() {
	log.Println("Hello world !")
	log.Println(Name, "- Version", Version)

	configPath := flag.String("config", "dlive.json", "Configuration file to use")
	flag.Parse()

	c, err := parseConfig(*configPath)
	if err != nil {
		log.Fatalf("Fatal error while parsing config: %s", err)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", c.DbUser, c.DbPass, c.DbHost, c.DbName))
	if err != nil {
		log.Fatalf("Database error: %s", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Database error: %s", err)
	}

	err = chat.Start(db)
	if err != nil {
		log.Fatalf("Fatal error when running server: %s", err)
	}

	go func() {
		err = events.Listen(db, ":8080")
		if err != nil {
			log.Fatalf("Can not start event server: %s", err)
		}
	}()
}
