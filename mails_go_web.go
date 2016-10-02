package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/veqryn/go-email/email"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Resolv  string `short:"r" long:"query-resolve" description:"Command to resolve a query into file path"`
}

func debug(message string, params ...interface{}) {
	if opts.Verbose {
		fmt.Printf(message+"\n", params...)
	}
}

func query_to_file(q string) string {
	var file_path string

	if opts.Resolv == "" {
		return q
	}

	res := fmt.Sprintf(opts.Resolv, q)
	cmd := exec.Command("sh", "-c", res)
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		file_path = lines[0]
	}

	debug("Query \"%s\" resolved as \"%s\" file path.", q, file_path)

	return file_path
}

func get_email_view(file_path string) string {
	var body string
	file, _ := ioutil.ReadFile(file_path)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)
	for _, part := range msg.MessagesAll() {
		mediaType, _, _ := part.Header.ContentType()
		switch mediaType {
		case "text/html":
			// by defaoult use html email
			body = string(part.Body)
		case "text/plain":
			// use plain body only if html not available
			if body == "" {
				body = string(part.Body)
			}
		}
	}

	return body
}

func download(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Not implemented yet...")
}

func view(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Query()["q"]) == 0 {
		io.WriteString(w, "Query not provided.")
		return
	}
	query := r.URL.Query()["q"][0]
	file_path := query_to_file(query)
	if _, err := os.Stat(file_path); os.IsNotExist(err) {
		io.WriteString(w, "Email not found.")
		return
	}

	body := get_email_view(file_path)
	// render body
	io.WriteString(w, body)
}

func main() {

	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	http.HandleFunc("/", view)
	http.HandleFunc("/attachment", download)

	fmt.Println("Starting server on http://localhost:6245")
	http.ListenAndServe("localhost:6245", nil)
}
