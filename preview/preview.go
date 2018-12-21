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
			max-width: 800px;
			height: 100%;
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
					</tr>
					<tr>
						<td class="field-name">
							To:
						</td>
						<td>
							{{.To}}
						</td>
					</tr>
					<tr>
						<td class="field-name">
							Date:
						</td>
						<td>
							{{.Date}}
						</td>
					</tr>
					<tr>
						<td class="field-name bd-b">
							Subject:
						</td>
						<td class="bd-b">
							{{.Subject}}
						</td>
					</tr>
				</table>
			</td>
		</tr>
		<tr>
			<td>
				<iframe frameborder="0" src="{{.MessageUrl}}">
				</iframe>
			</td>
		</tr>
	</table>`
}

func Template() string {
	return `
	<!doctype html>
	<html>
		<head>` + Css() + `</head>
		<body>` + Body() + `</body>
	</html>
	`
}
