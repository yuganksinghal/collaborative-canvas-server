package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type responseFlusher struct {
	http.ResponseWriter
	http.Flusher
}

type Message struct {
	Image string `json:"image"`
}

func (w *responseFlusher) Write(b []byte) (n int, err error) {
	if n, err = fmt.Fprintf(w.ResponseWriter, "data %s\n\n", b); err != nil {
		return
	}
	w.Flush()
	return
}

func sse(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming is not supported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Cache-Control", "no-cache")
		h(&responseFlusher{ResponseWriter: w, Flusher: f}, r)
	}
}

func main() {
	base64Image := ""
	http.HandleFunc("/canvas", sse(func(w http.ResponseWriter, r *http.Request) {
		for {
			fmt.Println("SSE EVENT: ", base64Image)
			time.Sleep(time.Second * 5)
			w.Write([]byte(base64Image))
		}
	}))

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			fmt.Println("issue reading body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var msg Message
		err = json.Unmarshal(b, &msg)
		if err != nil {
			fmt.Println("issue unmarshalling")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		base64Image = msg.Image
		w.Write([]byte("yeet"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
