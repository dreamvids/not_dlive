package stream

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/dreamvids/webm-info/webm"
	"github.com/gorilla/mux"
)

const (
	ChunckSize = 100000 // 100 ko
)

var (
	StreamsLastFrag    = make(map[string]int64)
	StreamsVideoHeader = make(map[string][]byte)
)

func HandlePush(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	StreamsLastFrag[vars["id"]] = 0

	f, err := os.Create(vars["id"] + ".webm")
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, ChunckSize)
	index := 0

	_, err = io.ReadFull(r.Body, buf)
	if err != nil {
		log.Println("Can not read from request body (push):", err)
		return
	}

	var doc webm.Document
	doc.Data = buf
	doc.Cursor = 0
	doc.Length = uint64(len(buf))

	headerData, err := webm.ReadHeader(&doc)
	if err != nil {
		log.Println("Invalid webm header (push):", err)
		return
	}

	StreamsVideoHeader[vars["id"]] = headerData
	index += len(headerData)

	for {
		if index != 0 {
			_, err = f.Write(buf[index:len(buf)])
			if err != nil {
				log.Println("Write to file:", err)
			}

			index = 0
		} else {
			n, err := io.ReadFull(r.Body, buf)
			if err != nil && err != io.ErrUnexpectedEOF {
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

		StreamsLastFrag[vars["id"]]++
	}

	log.Println("Push done for ID", vars["id"])
}

func HandlePullInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lastFrag := StreamsLastFrag[vars["id"]]

	if lastFrag >= 10 {
		lastFrag -= 10
	} else {
		lastFrag = -1
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(fmt.Sprintf("{\"last_fragment\": %d}", lastFrag)))
}

func HandlePull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	file, err := os.Open(id + ".webm")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	var doc webm.Document
	buf := make([]byte, ChunckSize)

	_, err = io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		log.Println("Read error:", err)
		return
	}

	doc.Cursor = 0
	doc.Length = uint64(len(buf))
	doc.Data = buf

	headerData, err := webm.ReadHeader(&doc)
	if err != nil {
		log.Println("Invalid webm header (push):", err)
		return
	}

	w.Write(headerData)
	fmt.Println("wrote", len(headerData), buf[0])

	/*for {
		_, err := io.ReadFull(file, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			log.Println("Read error:", err)
			break
		}

		doc.Data = append(doc.Data, buf...)
		doc.Length += uint64(len(buf))

		for {
			_, err := webm.ReadBlock(&doc)
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("Pull: can not read block:", err)
				continue
			}
		}
	}*/
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
