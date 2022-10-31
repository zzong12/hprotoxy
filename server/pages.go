package server

var WebPages = map[string]string{}

func init() {
	WebPages["/"] = PageIndex
	WebPages["/index.html"] = PageIndex
}

const (
	PageIndex = `
	<!DOCTYPE html>
	<html>
	  <head>
		<meta charset="utf-8">
		<title>My App</title>
		<style>
			.body {font-size: 12px;}
			#meta-table {border-collapse: collapse;border-style: solid;}
			#meta-table td {border-style: solid; border-collapse:collapse;padding: 5px;height: 10px;word-wrap: inherit;font-size: 14px;}
			#meta-table td textarea {height: 20px;width: 400px;}
			#meta-table header {background-color: #eee;font: bold;}
			#ctl-box {height: 20px;border: solid;padding: 5px;}
		</style>
	  </head>
	  <body>
		<div id="ctl-box">
			<form id="ctl-form" action="/do/upload" enctype="multipart/form-data" method="post">
				Place your file here:
				<input type="file" multiple="multiple" id="ctl-file" name="pbfile"/>
				<button type="button" id="ctl-upload" onclick="doUpload()">Upload</button>
				<button type="button" id="ctl-reload" onclick="doReload()">Reload</button>
			</form>
		</div>
		<div id="meta-box">
			<table id="meta-table">
				<header>
					<tr>
						<td>Filename</td>
						<td>MessageName</td>
						<td>MessageType</td>
						<td>Example</td>
					</tr>
				</header>
			</table>
		</div>
	  </body>
	  <script>
		var Ajax = {
			get: function (url, fn) {
				var xhr = new XMLHttpRequest();
				xhr.open('GET', url, true);
				xhr.setRequestHeader('Content-Type','application/x-www-form-urlencode;charset=utf-8');
				xhr.onreadystatechange = function () {
					if (xhr.readyState === 4 && xhr.status === 200) {
						fn.call(this, xhr.responseText)
					}
				};
				xhr.send()
			},
			post: function (url, data, fn) {
				var xhr = new XMLHttpRequest();
				xhr.open("POST", url, true);
				xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
				xhr.onreadystatechange = function() {
					if (xhr.readyState == 4 && (xhr.status == 200 || xhr.status == 304)) {
						fn.call(this, xhr.responseText);
					}
				};
				xhr.send(data);
			}
		};
		function doUpload(event) {
			var e = event || window.event;
			e.preventDefault();
			document.getElementById('ctl-form').submit();
		};
		function doReload(event) {
			var e = event || window.event;
			e.preventDefault();
			Ajax.get('/do/reload', function (data) {
				alert(data);
			})
		};
		(function() {
			Ajax.get('/st/meta', function (data) {
				var metaData = JSON.parse(data);
				var tbl = "<tbody>";
				metaData.forEach(function (item) {
					tbl += "<tr>";
					tbl += "<td>" + item.fileName + "</td>";
					tbl += "<td>" + item.msgName + "</td>";
					tbl += "<td>" + item.msgType + "</td>";
					tbl += "<td><textarea>" + item.example + "</textarea></td>";
					tbl += "</tr>";
				})
				tbl += "</tbody>";
				document.getElementById('meta-table').innerHTML += tbl;
			});
		})();
	  </script>
	</html>
	`
)