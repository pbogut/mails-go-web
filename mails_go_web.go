package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/veqryn/go-email/email"
)

func download(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Not implemented yet...")
}

func view(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()["q"]) == 0 {
		io.WriteString(w, "Query not provided.")
		return
	}
	query := r.URL.Query()["q"][0]
	if _, err := os.Stat(query); os.IsNotExist(err) {
		io.WriteString(w, "Email not found.")
		return
	}

	file, _ := ioutil.ReadFile(query)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)
	for _, part := range msg.MessagesAll() {
		mediaType, _, _ := part.Header.ContentType()
		switch mediaType {
		case "text/html":
			io.WriteString(w, string(part.Body))
		case "text/plain":
			io.WriteString(w, string(part.Body))
		}
	}
}

func main() {
	http.HandleFunc("/", view)
	http.HandleFunc("/attachment", download)

	fmt.Println("Sstarting server on http://localhost:6245")
	http.ListenAndServe("localhost:6245", nil)
}
