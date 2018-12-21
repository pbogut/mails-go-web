package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"./preview"

	"github.com/jessevdk/go-flags"
	"github.com/veqryn/go-email/email"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Resolv  string `short:"r" long:"query-resolve" description:"Command to resolve a query into file path"`
	Port    string `short:"p" long:"port" description:"Server port" default:"6245"`
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

func get_email_body(file_path string) string {
	var body string
	file, _ := ioutil.ReadFile(file_path)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)
	for _, part := range msg.MessagesAll() {
		mediaType, _, _ := part.Header.ContentType()
		switch mediaType {
		case "text/html":
			// by defaoult use html email
			html := string(part.Body)
			// open links in parent window
			body = strings.Replace(html, "<head>", `<head><base target="_parent">`, 1)
		case "text/plain":
			// use plain body only if html not available
			if body == "" {
				body = string(part.Body)
			}
		}
	}

	return body
}

func get_email_view(file_path string, url string) string {
	file, _ := ioutil.ReadFile(file_path)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)

	re := regexp.MustCompile("[^<]*<(.*)>")
	from := strings.ToLower(re.ReplaceAllString(msg.Header.From(), "$1"))
	data := []byte(from)
	hash := fmt.Sprintf("%x", md5.Sum(data))

	date, _ := msg.Header.Date()

	preview_html := preview.Template()

	m := map[string]interface{}{
		"EmailHash":  hash,
		"From":       html.EscapeString(msg.Header.From()),
		"To":         html.EscapeString(strings.Join(msg.Header.To(), ", ")),
		"Date":       date.Format("Mon, 2 Jan [2006-01-02 15:04:05]"),
		"Subject":    html.EscapeString(msg.Header.Subject()),
		"MessageUrl": url,
	}

	content := new(bytes.Buffer)
	templ, _ := template.New("preview").Parse(preview_html)
	templ.Execute(content, m)

	return content.String()
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

	var body string
	if len(r.URL.Query()["m"]) == 0 {
		body = get_email_view(file_path, r.URL.Path+"?m=true&"+r.URL.RawQuery)
	} else {
		body = get_email_body(file_path)
	}
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

	fmt.Println("Starting server on http://localhost:" + opts.Port)
	http.ListenAndServe("localhost:"+opts.Port, nil)
}
