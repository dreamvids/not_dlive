package stream

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dreamvids/webm-info/webm"
	"github.com/gorilla/mux"
)

const (
	ChunckSize = 100000 // 100 KB
)

func HandlePush(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	f, err := os.Create(vars["id"] + ".webm")
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, ChunckSize)

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

	log.Println("Push done for ID", vars["id"])
}

func HandlePullInfo(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Write([]byte(fmt.Sprintf("{\"last_fragment\": %d}", lastFrag)))
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

	buf, err := ioutil.ReadFile(id + ".webm")
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

	doc.Cursor = 0
	doc.Length = uint64(len(buf))
	doc.Data = buf

	headerData, err := webm.ReadHeader(&doc)
	if err != nil {
		log.Println("Invalid webm header (pull):", err)
		return
	}

	w.Write(headerData)
	flusher.Flush()

	for {
		clusterData, err := webm.ReadClusterData(&doc)
		if err != nil {
			log.Println("Invalid webm cluster (pull):", err)
			break
		}

		w.Write(clusterData)
		flusher.Flush()

		for {
			block, err := webm.ReadBlock(&doc)
			if err == webm.EndOfBlock {
				break
			} else if err != nil {
				log.Printf("Invalid webm block (pull): %s -- %d of %d\n", err, doc.Cursor, doc.Length)
				return
			}

			w.Write(block)
			flusher.Flush()
		}
	}
}

func HandlePullFrag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	file, err := os.Open(id + ".webm")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	frag, err := strconv.Atoi(vars["frag"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error:", err)
		return
	}

	start := int64(frag * ChunckSize)
	buf := make([]byte, ChunckSize)

	n, err := file.ReadAt(buf, start)
	if n != ChunckSize {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Error: Invalid bytes red from file (%d != expected %d)\n", n, ChunckSize)
		return
	}

	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ChunckSize))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(buf)
}
