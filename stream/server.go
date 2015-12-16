package stream

import (
	"fmt"
	"net/http"
	"os"
)

func Push(id string, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Push", id)

	file, err := os.Create(fmt.Sprintf("%s.webm", id))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := make([]byte, 10000) // 10kb

	for {
		n, err := r.Body.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}

		_, err = file.Write(buf[0:n])
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func Pull(id string, w http.ResponseWriter, r *http.Request) {

}
