package stream

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dreamvids/webm-info/webm"
	"github.com/gorilla/mux"
)

const (
	ChunckSize = 30000
)

var (
	Streaming = make(map[string]bool)
)

func HandlePush(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	f, err := os.Create(vars["id"] + ".webm")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initiating push for ID", vars["id"])

	buf := make([]byte, ChunckSize)
	Streaming[vars["id"]] = true

	for {
		n, err := io.ReadFull(r.Body, buf)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		_, err = f.Write(buf)
		if err != nil {
			log.Println("Write to file:", err)
		}

		if n != len(buf) {
			break
		}
	}

	Streaming[vars["id"]] = false
	log.Println("Push done for ID", vars["id"])
}

func HandlePull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Println("Can not setup response (flushder)")
		return
	}

	var doc webm.Document

	file, err := os.Open(id + ".webm")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Connection", "keep-alive")

	log.Println("Initiating pull for ID", vars["id"])

	ck := 0
	loadVideoChunk(&doc, file, &ck)

	headerData, err := webm.ReadHeader(&doc)
	if err != nil {
		log.Println("Invalid webm header (pull):", err)
		return
	}

	w.Write(headerData)
	flusher.Flush()

	lastClPos := uint64(0)
	for {
		lastClPos = doc.Cursor

		clusterData, err := webm.ReadCluster(&doc)
		if err != nil {
			if !Streaming[id] {
				log.Println("Pull done for ID", vars["id"])
				return
			}

			for {
				err = loadVideoChunk(&doc, file, &ck)
				if err != nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				break
			}

			doc.Cursor = lastClPos
			continue
		}

		w.Write(clusterData)
		flusher.Flush()
	}
}

func loadVideoChunk(doc *webm.Document, file *os.File, i *int) error {
	buf := make([]byte, ChunckSize)

	n, err := file.ReadAt(buf, int64((*i)*ChunckSize))
	if err != nil {
		if n <= 0 {
			return io.EOF
		}
	}

	doc.Data = append(doc.Data, buf...)
	doc.Length += ChunckSize
	*i++

	return nil
}
