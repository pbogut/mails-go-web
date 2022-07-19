package main

import (
	"bytes"
	"path"
	"crypto/md5"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/pbogut/mails-go-web/preview"

	"github.com/djimenez/iconv-go"
	"github.com/jessevdk/go-flags"
	"github.com/veqryn/go-email/email"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Resolv  string `short:"r" long:"query-resolve" description:"Command to resolve a query into file path"`
	Port    string `short:"p" long:"port" description:"Server port" default:"6245"`
	Host    string `short:"s" long:"host" description:"Server host" default:"localhost"`
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

func get_email_body(file_path string, query string) string {
	var body string
	msg := email_file_to_msg(file_path)
	re := regexp.MustCompile("<(img)([^>]*)(src)=\"cid:([^\"]*)\"([^>]*)>")
	for _, part := range msg.MessagesAll() {
		mediaType, extra, _ := part.Header.ContentType()
		switch mediaType {
		case "text/html":
			// by defaoult use html email
			html := string(part.Body)
			if len(extra["charset"]) > 0 {
				html = convert(html, extra["charset"], "UTF-8")
			}
			idx := strings.Index(body, "<!doctype")
			if idx == -1 || idx > 50 {
				html = "<!doctype html>\n<base target=\"_parent\">" + html
			}
			// open links in parent window
			body = strings.Replace(html, "<head>", `<head><base target="_parent">`, 1)
			body = re.ReplaceAllString(body, "<$1$2$3=\"/?q="+query+"&file=$4\"$5>")
		case "text/plain":
			// use plain body only if html not available
			if body == "" {
				body = string(part.Body)
				if len(extra["charset"]) > 0 {
					body = convert(body, extra["charset"], "UTF-8")
				}
			}
		}
	}

	return body
}

func get_email_view(file_path string, query string) string {
	msg := email_file_to_msg(file_path)

	from := strings.ToLower(extract_from_angle_brackets(msg.Header.From()))
	data := []byte(from)
	hash := fmt.Sprintf("%x", md5.Sum(data))

	date, _ := msg.Header.Date()

	preview_html := preview.Template()

	var parts []map[string]string
	for _, part := range msg.MessagesAll() {
		partType, disposition, _ := part.Header.ContentDisposition()
		if partType == "attachment" || partType == "inline" {
			fileName := html.UnescapeString(q_string_decode(disposition["filename"]))
			parts = append(parts, map[string]string{
				"Url":  "/?q=" + query + "&file=" + url.QueryEscape(fileName),
				"Name": fileName,
			})

		}
	}

	m := map[string]interface{}{
		"EmailHash": hash,
		"From":      html.EscapeString(q_string_decode(msg.Header.From())),
		"To":        html.EscapeString(q_string_decode(strings.Join(msg.Header.To(), ", "))),
		"Date":      date.Format("Mon, 2 Jan [2006-01-02 15:04:05]"),
		"Subject":   html.EscapeString(q_string_decode(msg.Header.Subject())),
		"Query":     query,
		"Parts":     parts,
	}

	content := new(bytes.Buffer)
	templ, _ := template.New("preview").Parse(preview_html)
	templ.Execute(content, m)

	return content.String()
}

func get_email_attachment(file_path string, attachment_name string) string {
	msg := email_file_to_msg(file_path)

	for _, part := range msg.MessagesAll() {
		_, disposition, _ := part.Header.ContentDisposition()
		candidates := []string{
			html.UnescapeString(q_string_decode(disposition["filename"])),
			part.Header.Get("X-Attachment-Id"),
			extract_from_angle_brackets(part.Header.Get("Content-Id")),
		}
		if contains(candidates, attachment_name) {
			return string(part.Body)
		}
	}

	return file_path + attachment_name
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

	if len(r.URL.Query()["eml"]) > 0 {
		w.Header().Add("Content-Disposition", "attachment; filename=\""+path.Base(file_path)+".eml\"")
		raw_content, _ := ioutil.ReadFile(file_path)
		body = string(raw_content)
	} else if len(r.URL.Query()["file"]) > 0 {
		attachment_name, _ := url.QueryUnescape(r.URL.Query()["file"][0])
		w.Header().Add("Content-Disposition", "attachment; filename=\""+attachment_name+"\"")
		body = get_email_attachment(file_path, attachment_name)
	} else if len(r.URL.Query()["raw"]) == 0 {
		body = get_email_view(file_path, query)
	} else {
		body = get_email_body(file_path, query)
	}
	// render body
	io.WriteString(w, body)
}

func email_file_to_msg(file_path string) *email.Message {
	file, _ := ioutil.ReadFile(file_path)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)

	return msg
}

func q_string_decode(name string) string {
	re := regexp.MustCompile("=\\?([a-zA-Z0-9_\\-]*?)\\?Q\\?(.*?)\\?=")

	for _, part := range re.FindAllStringSubmatch(name, -1) {
		full := part[0]
		charset := part[1]
		text := part[2]

		re = regexp.MustCompile("=([A-F0-9][A-F0-9])")
		newValue := re.ReplaceAllString(text, "%$1")
		newValue, _ = url.PathUnescape(newValue)
		newValue = convert(newValue, charset, "UTF-8")

		name = strings.Replace(name, full, newValue, -1)
	}

	return name
}

func extract_from_angle_brackets(text string) string {
	re := regexp.MustCompile("[^<]*<(.*)>")
	return strings.ToLower(re.ReplaceAllString(text, "$1"))
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if strings.ToLower(a) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func convert(text string, src string, dest string) string {
	result, err := iconv.ConvertString(text, src, dest)
	if err != nil {
		return text
	}
	return result
}

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		return
	}

	http.HandleFunc("/", view)

	fmt.Println("Starting server on http://" + opts.Host + ":" + opts.Port)
	http.ListenAndServe(opts.Host + ":"+opts.Port, nil)
}
