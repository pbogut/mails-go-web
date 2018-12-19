package main

import (
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

	"github.com/jessevdk/go-flags"
	"github.com/veqryn/go-email/email"
)

var opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Resolv  string `short:"r" long:"query-resolve" description:"Command to resolve a query into file path"`
  Port string `short:"p" long:"port" description:"Server port" default:"6245"`
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

func get_email_view(file_path string, url string) string {
	file, _ := ioutil.ReadFile(file_path)
	reader := strings.NewReader(string(file))
	msg, _ := email.ParseMessage(reader)

	body := `<!doctype html>
	<html>
		<head>
			<style>
			body, html {
				height: 100%%;
				margin: 0;
				padding: 0;
				overflow: hidden;
			}
			.container {
				margin: auto;
				max-width: 800px;
				height: 100%%;
				border: 1px solid black;
				box-sizing: border-box;
			}
			.header {
				font-size: 14px;
				height: 84px;
				box-sizing: border-box;
			}
			.header img {
				float: left;
			}
			.header table {
				width: 100%%;
			}
			.header td.gravatar {
				width: 80px;
			}
			.header td.field-name {
				padding-right: 5px;
				width: 70px;
				font-weight: bold;
				text-align: right;
			}
			.header td.bd-r {
				border-right: 1px solid black;
			}
			.header td.bd-b {
				border-bottom: 1px solid black;
			}
			iframe {
				height: calc(100%% - 84px);
				width: 100%%;
			}
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
				<table cellspacing="0" border="0">
					<tr>
						<td rowspan="4" class="bd-b bd-r gravatar">
							<img src="http://www.gravatar.com/avatar/%x?s=80&d=identicon" alt="" />
						</td>
						<td class="field-name">
							From:
						</td>
						<td class="bd-r">
							%s
						</td>
					</tr>
					<tr>
						<td class="field-name">
							To:
						</td>
						<td class="bd-r">
							%s
						</td>
					</tr>
					<tr>
						<td class="field-name">
							Date:
						</td>
						<td class="bd-r">
							%s
						</td>
					</tr>
					<tr>
						<td class="field-name bd-b">
							Subject:
						</td>
						<td class="bd-b bd-r">
							%s
						</td>
					</tr>
				</table>
				</div>
				<iframe frameborder="0" src="%s">
				</iframe>
			</div>
		</body>
	</html>
	`

	re := regexp.MustCompile("[^<]*<(.*)>")
	from := strings.ToLower(re.ReplaceAllString(msg.Header.From(), "$1"))
	data := []byte(from)
	hash := md5.Sum(data)

	date, _ := msg.Header.Date()

	return fmt.Sprintf(
		body,
		hash,
		html.EscapeString(msg.Header.From()),
		msg.Header.To(),
		date.Format("Mon, 2 Jan [2006-01-02 15:04:05]"),
		html.EscapeString(msg.Header.Subject()),
		url,
	)
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
	http.ListenAndServe("localhost:" + opts.Port, nil)
}
