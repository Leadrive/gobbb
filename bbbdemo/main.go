package main

import (
	"flag"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"net/url"
	"os"

	"github.com/sdgoij/gobbb"
)

var (
	flagHttpAddr     = flag.String("http.addr", ":8080", "HTTP service address (e.g., ':8080')")
	flagServerURL    = flag.String("server.url", "", "BigBlueButton API URL to connect")
	flagServerSecret = flag.String("server.secret", "", "BigBlueButton API secret")
	flagLogOutput    = flag.String("log.output", "", "Logfile, 'syslog' or 'nil'")

	templates *template.Template
	b3        bbb.BigBlueButton
)

type _error string

func (e _error) Error() string {
	return string(e)
}

func init() {

	templates = template.Must(template.New("index.html").Parse(pageIndexTemplate))
	// template.Must(templates.New("connect.html").Parse(pageConnectTemplate))
	template.Must(templates.New("create.html").Parse(pageCreateTemplate))
	template.Must(templates.New("info.html").Parse(pageInfoTemplate))
	template.Must(templates.New("join.html").Parse(pageJoinTemplate))

	http.HandleFunc("/", PgIndex)
	http.HandleFunc("/connect", PgConnect)
	http.HandleFunc("/create", PgCreate)
	http.HandleFunc("/info", PgInfo)
	http.HandleFunc("/join", PgJoin)

	flag.Parse()
}

func main() {
	if *flagLogOutput != "" {
		logging(*flagLogOutput)
	}
	b3, _ = bbb.New(*flagServerURL, *flagServerSecret)
	log.Fatal(http.ListenAndServe(*flagHttpAddr, Log(http.DefaultServeMux)))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func logging(output string) {
	var (
		writer io.Writer
		err    error
	)
	switch output {
	case "nil", "-":
		writer, err = ioutil.Discard, nil
	case "syslog":
		writer, err = syslog.New(syslog.LOG_NOTICE, "bbbdemo")
		log.SetFlags(log.Lshortfile)
	default:
		flags := os.O_APPEND | os.O_WRONLY
		if _, err := os.Stat(output); nil != err {
			flags |= os.O_CREATE
		}
		writer, err = os.OpenFile(output, flags, 0644)
	}
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Setting log output to:", output)
	log.SetOutput(writer)
	log.Println("Started")
}

func PgIndex(w http.ResponseWriter, req *http.Request) {
	data := struct {
		ServerVersion string
		ServerURL     string
		Meetings      []*bbb.Meeting
	}{
		b3.ServerVersion(),
		b3.Url.String(),
		b3.Meetings(),
	}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PgConnect(w http.ResponseWriter, req *http.Request) {
	if "POST" == req.Method {
		apiurl, secret := *flagServerURL, *flagServerSecret
		if v := req.PostFormValue("apiurl"); v != "" {
			apiurl = v
		}
		if v := req.PostFormValue("secret"); v != "" {
			secret = v
		}
		b3, _ = bbb.New(apiurl, secret)
		log.Println("New bbb settings:", b3.Url.String(), b3.Secret)
	}
	req.Method = "GET"
	http.Redirect(w, req, "/", http.StatusFound)
}

func PgCreate(w http.ResponseWriter, req *http.Request) {
	if "POST" == req.Method {
		options := &bbb.CreateOptions{
			Record: req.FormValue("record") == "1",
			Name:   req.FormValue("name"),
		}
		if m, err := b3.Create(req.FormValue("id"), options); nil != err {
			http.Error(w, err.Error(), http.StatusTeapot)
			return
		} else {
			req.Method = "GET"
			q := url.Values{"id": {m.Id}}
			http.Redirect(w, req, "/info?"+q.Encode(), http.StatusFound)
		}
	} else {
		if err := templates.ExecuteTemplate(w, "create.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func PgInfo(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	for _, meeting := range b3.Meetings() {
		if meeting.Id == id {
			log.Printf("%#v", meeting)
			if meeting, err := b3.MeetingInfo(id, meeting.ModeratorPW); nil == err {
				if err := templates.ExecuteTemplate(w, "info.html", meeting); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			} else {
				http.Error(w, err.Error(), http.StatusTeapot)
			}
			return
		}
	}
	http.Error(w, "Meeting by id '"+id+"' not found!", http.StatusTeapot)
}

func PgJoin(w http.ResponseWriter, req *http.Request) {
	i, m, u := req.FormValue("id"), req.FormValue("m"), req.FormValue("u")
	if "" == u {
		if err := templates.ExecuteTemplate(w, "join.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	for _, meeting := range b3.Meetings() {
		if meeting.Id == i {
			password := meeting.AttendeePW
			if m == "1" {
				password = meeting.ModeratorPW
			}
			joinURL := b3.JoinURL(u, i, password, bbb.EmptyOptions)
			// if err := templates.ExecuteTemplate(w, "join.html", joinURL); err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }
			http.Redirect(w, req, joinURL, http.StatusFound)
		}
	}
	http.Error(w, "Meeting by id '"+i+"' not found!", http.StatusTeapot)
}

const pageIndexTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
    <title>BigBlueButton Server Admin</title>
  </head>
  <body>
    <div id="main">
      <div id="header">
      	<div id="connect">
          <form action="/connect" method="post">
            <input type="text" name="apiurl" {{if .ServerURL}}value="{{.ServerURL}}"{{end}}/>
            <input type="text" name="secret"/>
            <input type="submit"/>
          </form>
        </div>
        <div id="server-version">ServerVersion: {{.ServerVersion}}</div>
      </div>
      <hr/>
      <div id="meetings">
        <dl>
        {{range .Meetings}}
          <dt>{{.Id}}</dt>
          <dt>{{.Name}}</dt>
          <dd><a href="/info?id={{.Id}}">Info</a></dd>
          <dd><a href="/join?id={{.Id}}">Join</a></dd>
          <dd><a href="/join?id={{.Id}}&m=1">Join(moderator)</a></dd>
      	{{else}}
      		<dd>No meetings found<dd>
      		<dt>Try <a href="/create">create</a> one.</dt>
      	{{end}}
      	</dl>
      </div>
    </div>
  </body>
</html>`

// const pageConnectTemplate = `
// <!DOCTYPE html>
// <html lang="en">
//   <head>
//     <meta charset="utf-8"/>
// 	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
//     <title>BigBlueButton Server Admin</title>
//   </head>
//   <body>
//     <div id="main">
//       <form action="/create" method="post">
//         <!-- ... -->
//       </form>
//     </div>
//   </body>
// </html>`

const pageCreateTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
    <title>BigBlueButton Server Admin</title>
  </head>
  <body>
    <div id="main">
      <form action="/create" method="post">
        <input type="text" name="id"/>
        <input type="text" name="name"/>
        <input type="checkbox" name="record" value="1"/>
        <input type="submit"/>
      </form>
    </div>
  </body>
</html>`

const pageInfoTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
    <title>BigBlueButton Server Admin</title>
    <style type="text/css">
      dl {
        margin-bottom: 50px;
      }
      dl dt {
        background: #5f9be3;
        color: #fff;
        float: left;
        font-weight: bold;
        margin-right: 10px;
        padding: 5px;
        width: 150px;
      }
      dl dd {
        margin: 2px 0;
        padding: 5px 0;
      }
    </style>
  </head>
  <body>
    <div id="main">
      <dl>
        <dt>Id:</dt>
        <dd>{{.Id}}</dd>

        <dt>Name:</dt>
        <dd>{{.Name}}</dd>

        <dt>Created:</dt>
        <dd>{{.CreateTime}}</dd>

        <dt>Attendee:</dt>
        <dd>{{.AttendeePW}}</dd>

        <dt>Moderator:</dt>
        <dd>{{.ModeratorPW}}</dd>

        <dt>Running:</dt>
        <dd>{{.Running}}</dd>
      </dl>
      <hr/>
      <div>
        <a href="/">Back</a> | <a href="/join?id={{.Id}}">Join</a> | <a href="/end?id={{.Id}}">End</a>
      </div>
    </div>
  </body>
</html>`

// const pageJoinTemplate = `
// <!DOCTYPE html>
// <html lang="en">
//   <head>
//     <meta charset="utf-8"/>
// 	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
//     <title>BigBlueButton Server Admin</title>
//   </head>
//   <body>
//     <script type="text/javascript">window.location="{{.}}";</script>
//   </body>
// </html>`

const pageJoinTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
	  <meta name="viewport" content="initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, width=device-width" />
    <title>BigBlueButton Server Admin</title>
  </head>
  <body>
    <script type="text/javascript">
      var u = window.prompt("Please enter your name");
      window.location+="&u="+encodeURIComponent(u);
    </script>
  </body>
</html>`
