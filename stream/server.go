package stream

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/quadrifoglio/go-webm"
)

type Stream struct {
	Header      []byte
	Cluster     []byte
	LastCluster int
}

var (
	Streams = make(map[string]*Stream)
	Mutex   = &sync.Mutex{}
)

func Push(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	buf := make([]byte, 1000000) // 1MB (max frame size)

	Mutex.Lock()
	Streams[id] = &Stream{make([]byte, 1), make([]byte, 1), 0}
	Mutex.Unlock()

	file, err := os.OpenFile(id+".webm", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	for {
		n, err := r.Body.Read(buf)
		if err != nil {
			fmt.Println(err)
			Streams[id].Cluster = nil
			break
		}

		if n <= 0 {
			continue
		}

		doc := webm.InitDocument(buf)
		el, err := doc.ParseElement()
		if err != nil {
			fmt.Println("WebM:", err, "at", doc.Cursor)
			continue
		}

		fmt.Printf("Element: %s (%x)\n", el.Name, el.ID)

		if el.ID == webm.ElementEBML.ID {
			Mutex.Lock()

			if n != len(Streams[id].Header) {
				Streams[id].Header = make([]byte, n)
			}

			copy(Streams[id].Header, buf[0:n])
			file.Write(buf[0:n])

			Mutex.Unlock()
		}
		if el.ID == webm.ElementCluster.ID {
			Mutex.Lock()

			if n != len(Streams[id].Cluster) {
				Streams[id].Cluster = make([]byte, n)
			}

			copy(Streams[id].Cluster, buf[0:n])
			file.Write(buf[0:n])
			Streams[id].LastCluster++

			Mutex.Unlock()
		}
	}

	Mutex.Lock()
	Streams[id] = nil
	Mutex.Unlock()
}

func Pull(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Pull", id)

	Mutex.Lock()
	defer Mutex.Unlock()

	if Streams[id] == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("wtf")
		return
	}

	var lastSent = Streams[id].LastCluster

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Connection", "keep-alive")

	w.Write(Streams[id].Header)
	flusher.Flush()
	Mutex.Unlock()

	for {
		Mutex.Lock()
		if Streams[id] == nil {
			break
		}

		cl := Streams[id].Cluster
		if cl == nil {
			break
		}

		if lastSent < Streams[id].LastCluster {
			_, err := w.Write(cl)
			if err != nil {
				return
			}

			flusher.Flush()
			lastSent++
		}

		Mutex.Unlock()
		time.Sleep(1 * time.Millisecond)
	}
}
