package main

import (
	//"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	//"github.com/dreamvids/dlive/chat"
	"github.com/dreamvids/dlive/stream"
)

const (
	Name    string = "DreamVids Live Server"
	Version string = "1.0"
)

type serverConfig struct {
	HttpPort int `json:"http-port"`

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
	configPath := flag.String("config", "dlive.json", "Configuration file to use")
	logPath := flag.String("log", "", "Log file to use")
	flag.Parse()

	if len(*logPath) > 0 {
		f, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Println("Can not open log file:", err)
		} else {
			log.SetOutput(f)
		}
		defer f.Close()
	}

	c, err := parseConfig(*configPath)
	if err != nil {
		log.Fatalf("Fatal error while parsing config: %s", err)
	}

	/*log.Println("Connecting to database...")
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", c.DbUser, c.DbPass, c.DbHost, c.DbName))
	if err != nil {
		log.Fatalf("Database error: %s", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Database error: %s", err)
	}
	log.Println("Connected to database !")

	chat.Init(db)*/

	r := mux.NewRouter()
	//r.HandleFunc("/", chat.HandleWebsocket)

	r.HandleFunc("/stream/push/{id}", stream.HandlePush).Methods("POST")
	r.HandleFunc("/stream/pull/{id}", stream.HandlePull).Methods("GET")

	r.HandleFunc("/stream/player", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "stream_player.html")
	})

	http.Handle("/", r)
	log.Println("Listening on port", c.HttpPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.HttpPort), nil))
}
