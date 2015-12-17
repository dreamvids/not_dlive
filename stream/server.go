package stream

import (
	"fmt"
	"net/http"
	"time"

	"github.com/quadrifoglio/go-webm"
)

type Stream struct {
	Header  []byte
	Cluster []byte
}

var (
	Streams = make(map[string]*Stream)
)

func Push(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	buf := make([]byte, 1000000) // 1MB (max frame size)
	Streams[id] = &Stream{nil, nil}

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
			Streams[id].Cluster = nil
			break
		}

		fmt.Printf("Element: %s (%x)\n", el.Name, el.ID)

		if el.ID == webm.ElementEBML.ID {
			Streams[id].Header = buf[0:n]
		}
		if el.ID == webm.ElementCluster.ID {
			Streams[id].Cluster = buf[0:n]
		}
	}

	Streams[id] = nil
}

func Pull(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Pull", id)

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

	w.Write(Streams[id].Header)
	flusher.Flush()

	var lastLen = 0

	for {
		if Streams[id] == nil {
			break
		}

		cl := Streams[id].Cluster
		if cl == nil {
			break
		}

		if len(cl) != lastLen {
			w.Write(cl)
			flusher.Flush()

			lastLen = len(cl)
		}

		time.Sleep(1 * time.Millisecond)
	}
}
