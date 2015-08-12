package stream

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const (
	ChunckSize = 1048576
)

func HandlePush(w http.ResponseWriter, r *http.Request) {
	log.Printf("REQ - %s %v\n", r.Method, r.Header)

	vars := mux.Vars(r)

	f, err := os.Create(vars["id"] + ".webm")
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, ChunckSize)
	for {
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

	log.Println("done")
}

func HandlePullInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	file, err := os.Open(vars["id"] + ".webm")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	info, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	size := info.Size()
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
}

func HandlePull(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	file, err := os.Open(vars["id"] + ".webm")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	stat, err := file.Stat()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("Error:", err)
		return
	}

	totalSize := stat.Size()

	rangeStr := r.Header.Get("Range")
	if len(rangeStr) > 0 && strings.Contains(rangeStr, "=") {
		parts := strings.Split(rangeStr, "=")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rng := parts[1]
		parts = strings.Split(rng, "-")

		var min int64
		var max int64

		if len(parts) == 2 && len(parts[1]) > 0 {
			minI, err1 := strconv.Atoi(parts[0])
			maxI, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Println("Parse int error:", err1, err2)
				return
			}

			min = int64(minI)
			max = int64(maxI)
		} else {
			minI, err := strconv.Atoi(parts[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Println("Parse int error:", err)
				return
			}

			min = int64(minI)
			max = totalSize
		}

		if min < 0 || max < 0 || min > totalSize || max > totalSize || min >= max {
			log.Printf("Out of range: [%d-%d] - Max=%d\n", min, max, totalSize)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		buf := make([]byte, max-min)

		n, err := file.ReadAt(buf, min)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println("Error:", err)
			return
		}

		log.Printf("Serving [%d;%d] - len=%d\n", min, max, n)

		w.WriteHeader(http.StatusPartialContent)
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", n))
		w.Header().Set("Content-Type", "video/webm")
		w.Write(buf)
	} else {
		http.ServeFile(w, r, vars["id"]+".webm")
	}
}
