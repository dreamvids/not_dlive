package stream

import (
	"fmt"
	"net/http"
	"os"
	"time"

	//"github.com/hpcloud/tail"
	"github.com/quadrifoglio/go-mkv"
)

func Push(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	file, err := os.OpenFile(id+".webm", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	doc := mkv.InitDocument(r.Body)
	doc.ParseAll(func(el mkv.Element) {
		file.Write(el.Bytes)
	})
}

func Pull(id string, w http.ResponseWriter, r *http.Request) {
	var index int64 = 0

	flusher, _ := w.(http.Flusher)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Connection", "keep-alive")

	for {
		file, err := os.OpenFile(id+".webm", os.O_RDONLY, 0600)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Println(err)
			return
		}

		file.Seek(index, 0)

		doc := mkv.InitDocument(file)

		for {
			ti := index

			el, err := doc.ParseElement()
			if err != nil {
				index = ti
				break
			}

			w.Write(el.Bytes)
			flusher.Flush()
			index += int64(len(el.Bytes))
		}

		file.Close()
		time.Sleep(1 * time.Millisecond)
	}
}
