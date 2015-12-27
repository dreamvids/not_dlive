package stream

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/quadrifoglio/go-mkv"
)

var (
	header bytes.Buffer
	cls    bytes.Buffer
)

func Push(id string, notif chan int, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	doc := mkv.InitDocument(r.Body)
	headed := false

	for {
		el, err := doc.ParseElement()
		if err != nil {
			fmt.Println(err)
			break
		}

		if !headed {
			header.Write(el.Bytes)
		}

		if el.ID == mkv.ElementCluster.ID {
			if !headed {
				headed = true
			}

			_, err := doc.GetElementContent(&el)
			if err != nil {
				fmt.Println(err)
				break
			}

			cls.Reset()
			cls.Write(el.Bytes)

			select {
			case notif <- 1:
			default:
			}
		}
	}
}

func Pull(id string, notif chan int, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Pull", id)

	flusher, _ := w.(http.Flusher)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Connection", "keep-alive")

	w.Write(header.Bytes())

	for {
		<-notif
		w.Write(cls.Bytes())
		flusher.Flush()
	}

}
