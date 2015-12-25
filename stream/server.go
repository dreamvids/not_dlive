package stream

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/quadrifoglio/go-mkv"
)

type Stream struct {
	Header []byte
	Stream *io.ReadCloser
}

var (
	Streams = make(map[string]*Stream)
	Mutex   = &sync.Mutex{}
)

func Push(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	Mutex.Lock()
	Streams[id] = &Stream{make([]byte, 1), nil}

	file, err := os.OpenFile(id+".mkv", os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
		Mutex.Unlock()
		return
	}

	defer file.Close()
	Mutex.Unlock()

	doc := mkv.InitDocument(r.Body)
	doc.ParseAll(func(el mkv.Element) {
		// TODO: Stream the clusters to the clients
		fmt.Println("Element %s\n", el.Name)
	})

	Mutex.Lock()
	Streams[id] = nil
	Mutex.Unlock()
}

func Pull(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Pull", id)

	Mutex.Lock()

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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Connection", "keep-alive")

	w.Write(Streams[id].Header)
	flusher.Flush()
}
