package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nu7hatch/gouuid"
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
			var handlerFunc WsEventHandlerFunc
			var txid string

			if err := json.Unmarshal(src, &event); nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}
			responder := uhDefaultResponder
			switch event.Event {
			case "create":
				txid = addEventId(&event)
				handlerFunc = HandleCreate
				responder = uhMkIdResponder(txid)
			case "joinURL":
				handlerFunc = HandleJoinURL
			case "end":
				txid = addEventId(&event)
				handlerFunc = HandleEnd
				responder = uhMkIdResponder(txid)
			case "running":
				handlerFunc = HandleIsMeetingRunning
			case "info":
				handlerFunc = HandleMeetingInfo
			case "meetings":
				handlerFunc = HandleMeetings
			case "recordings":
				handlerFunc = HandleRecordings
			case "recordings.publish":
				txid = addEventId(&event)
				handlerFunc = HandlePublishRecordings
				responder = uhMkIdResponder(txid)
			case "recordings.delete":
				txid = addEventId(&event)
				handlerFunc = HandleDeleteRecordings
				responder = uhMkIdResponder(txid)
			default:
				handlerFunc = func(_ *Client, ev WsEvent) error {
					return _error("Unhandled event '" + ev.Event + "'")
				}
			}

			client := &Client{address: req.RemoteAddr, b3: b3, events: make(chan WsEvent, 1)}
			handler.AddClient(client)
			defer handler.RemoveClient(client)

			if err := handlerFunc(client, event); nil != err {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println(err)
				return
			}

			response, err := responder(client.events)
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

func addEventId(ev *WsEvent) string {
	if id, err := uuid.NewV4(); nil != err {
		panic(err.Error())
	} else {
		txid := id.String()
		ev.Data["__txid"] = txid
		return txid
	}
}

type uhResponderFunc func(chan WsEvent) ([]byte, error)

func uhDefaultResponder(events chan WsEvent) ([]byte, error) {
	return json.Marshal(<-events)
}

func uhMkIdResponder(id string) uhResponderFunc {
	return func(events chan WsEvent) ([]byte, error) {
		for {
			ev := <-events
			if txid, t := ev.Data["__txid"]; t && txid == id {
				delete(ev.Data, "__txid")
				return json.Marshal(ev)
			}
		}
		return nil, nil
	}
}
