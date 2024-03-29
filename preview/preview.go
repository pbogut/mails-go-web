package preview

func Css() string {
	return `
	<style>
		body, html {
			height: 100%;
			margin: 0;
			padding: 0;
			overflow: hidden;
		}
		.container {
			margin: auto;
			width: 800px;
			height: 100%;
			border: 1px solid black;
			box-sizing: border-box;
		}
		.attachments {
			height: 1px;
			border-top: 1px solid black;
		}
		.attachments a {
			display: inline-block;
			font-size: 14px;
			color: black;
			text-decoration: none;
			margin: 5px;
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
			width: 100%;
			height: 100%;
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
			height: 100%;
			width: 100%;
		}
		@media (prefers-color-scheme: dark) {
			body {
				background-color: #333;
				color: #eee;
			}
			a {
				color: lightblue;
			}
			.header td.bd-r {
				border-right: 1px solid #eee;
			}
			.header td.bd-b {
				border-bottom: 1px solid #eee;
			}
			.container {
				border: 1px solid #eee;
			}
			.attachments {
				border-top: 1px solid #eee;
			}
			iframe {
				background-color: #ccc;
			}
		}
	</style>
	`
}

func Body() string {
	return `
	<table class="container" cellspacing="0" border="0">
		<tr>
			<td class="header">
				<table cellspacing="0" border="0">
					<tr>
						<td rowspan="4" class="bd-b bd-r gravatar">
							<img src="http://www.gravatar.com/avatar/{{.EmailHash}}?s=80&d=identicon" alt="" />
						</td>
						<td class="field-name">
							From:
						</td>
						<td>
							{{.From}}
						</td>
						<td style="text-align: right">
							[<a href="/?q={{.Query}}&eml=true" target="_blank">eml</a>]
						</td>
					</tr>
					<tr>
						<td class="field-name">
							To:
						</td>
						<td colspan="2">
							{{.To}}
						</td>
					</tr>
					<tr>
						<td class="field-name">
							Date:
						</td>
						<td colspan="2">
							{{.Date}}
						</td>
					</tr>
					<tr>
						<td class="field-name bd-b">
							Subject:
						</td>
						<td colspan="2" class="bd-b">
							{{.Subject}}
						</td>
					</tr>
				</table>
			</td>
		</tr>
		<tr>
			<td>
				<iframe frameborder="0" src="/?q={{.Query}}&raw=true">
				</iframe>
			</td>
		</tr>
		{{if .Parts}}
		<tr>
			<td class="attachments">
				{{range .Parts}}
				<a href="{{.Url}}">[{{.Name}}]</a>
				{{end}}
			</td>
		</tr>
		{{end}}
	</table>`
}

func Scripts() string {
	return `
	<script>
		var iframe = document.getElementsByTagName('iframe')[0]
		document.onkeydown = function(ev) {
			var x=0, y=0
			if (ev.key == "ArrowUp") {
				y = -100
			} else if (ev.key == "ArrowDown") {
				y = 100
			} else if (ev.key == "ArrowLeft") {
				x = -100
			} else if (ev.key == "ArrowRight") {
				x = 100
			}
			iframe.contentWindow.scrollBy(x,y)
		}

	</script>
	`
}

func Template() string {
	return `
	<!doctype html>
	<html>
		<head>` + Css() + `</head>
		<body>` + Body() + Scripts() + `</body>
	</html>
	`
}
