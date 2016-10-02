# Mails Go Web

[![Project Status: Active - The project has reached a stable, usable state and is being actively developed.](http://www.repostatus.org/badges/latest/active.svg)](http://www.repostatus.org/#active)

HTTP server for rendering emails from Maildir files.

This go application creates HTTP server and renders emails based on the query.
By default, a query is a path to the email file. However, `--query-resolve` command
can be used to resolve different query type to actual file.

## But why?

I'm a mutt user, but often I need to view HTML emails in the browser.
Instead of saving the email to the file and point it in the browser
I wanted to make it simpler. At first, I was using mutt macro to save the email
and open file in the browser. I've used this macro from attachments view:

```
macro attach \ci '<shell-escape>rm -rf /tmp/.mutt-tmp/;mkdir /tmp/.mutt-tmp<enter><save-entry><kill-line>/tmp/.mutt-tmp/tmp-file.html<enter><shell-escape>rifle /tmp/.mutt-tmp/*<enter>'
```

But after opening the second email and refreshing tab with first one it was gone.
I wanted something more permanent, something that can work with many emails opened
and with an ability to bookmark email if I feel doing so. That's how this project
came to live.

### How to use it with mutt?

I'm using mutt macro to pipe email to bash script. The bash script is extracting
message id and it's opening the browser with URL to the mails-go-web server and
with id as a query.
*Notmuch* is used to find email by id and return its file path.

Mutt macro:
```
macro index,pager \ci "<pipe-message>mail-to-web.sh<return>" "html view in browser"
```
`mail-to-web.sh` can be found in scripts directory.

Mails go web command:
```
./mails-go-web -r "notmuch search --output=files id:%s"
```

That's it. When I'm pressing `Ctrl-I` email pops up in the browser.
Because I'm using Message-ID, the same link will work even if email file was moved.

## Contribution

Always welcome.

## License

MIT License;
