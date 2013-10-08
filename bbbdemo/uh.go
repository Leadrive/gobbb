package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sdgoij/gobbb"
)

func init() {
	http.HandleFunc("/uh", func(w http.ResponseWriter, req *http.Request) {
		if "POST" == req.Method {
			query := req.URL.Query()

			b3, err := bbb.New(query.Get("url"), query.Get("secret"))
			if nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("bbb.New:", err)
				return
			}

			src, err := ioutil.ReadAll(req.Body)
			log.Println(string(src), err)
			if nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
			defer req.Body.Close()

			var event WsEvent
			var handler WsEventHandlerFunc

			if err := json.Unmarshal(src, &event); nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
			switch event.Event {
			case "create":
				handler = HandleCreate
			case "joinURL":
				handler = HandleJoinURL
			case "end":
				handler = HandleEnd
			case "running":
				handler = HandleIsMeetingRunning
			case "info":
				handler = HandleMeetingInfo
			case "meetings":
				handler = HandleMeetings
			default:
				handler = func(_ *Client, ev WsEvent) error {
					return _error("Unhandled event '" + ev.Event + "'")
				}
			}

			client := &Client{address: req.RemoteAddr, b3: b3, events: make(chan WsEvent, 1)}

			if err := handler(client, event); nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}

			response, err := json.Marshal(<-client.events)
			if nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}

			if _, err := w.Write(response); nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
		}
	})
}
