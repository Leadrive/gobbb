package main

import (
	"io"
	"log"
	"net/http"
	"sync"

	"code.google.com/p/go.net/websocket"
	"github.com/sdgoij/gobbb"
)

func HandleConnect(c *Client, event WsEvent) error {
	url, secret := "", ""
	if u, t := event.Data["url"]; t && nil != u {
		url = u.(string)
	}
	if s, t := event.Data["secret"]; t && nil != s {
		secret = s.(string)
	}
	b3, err := bbb.New(url, secret)
	ev := WsEvent{"connected", WsEventData{
		"status":  "success",
		"version": "",
	}}
	if err == nil {
		if version := b3.ServerVersion(); "" == version {
			ev.Data["status"] = "failure"
		} else {
			ev.Data["version"] = version
			c.b3 = b3
		}
	}
	ev.Data["error"] = err.Error()
	c.events <- ev
	return err
}

func HandleCreate(c *Client, event WsEvent) error {
	id := ""
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}

	var response WsEvent

	if m, err := c.b3.Create(id, bbb.EmptyOptions); nil != err {
		response = WsEvent{"create.fail", WsEventData{"error": err.Error()}}
	} else {
		response = WsEvent{"create.success", WsEventData{
			"id":          m.Id,
			"created":     m.CreateTime.Unix(),
			"attendeePW":  m.AttendeePW,
			"moderatorPW": m.ModeratorPW,
			"forcedEnd":   m.ForcedEnd,
		}}
	}
	c.events <- response
	return nil
}

func HandleJoinURL(c *Client, event WsEvent) error {
	name, id, password := "", "", ""
	if v, t := event.Data["name"]; t && nil != v {
		name = v.(string)
	}
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}
	if v, t := event.Data["password"]; t && nil != v {
		password = v.(string)
	}
	c.events <- WsEvent{"joinURL", WsEventData{
		"url": c.b3.JoinURL(name, id, password, bbb.EmptyOptions),
	}}
	return nil
}

func HandleEnd(c *Client, event WsEvent) error { return nil }

var handler *WsEventHandler = &WsEventHandler{
	h: map[string]WsEventHandlerFunc{
		"connect": HandleConnect,
		"create":  HandleCreate,
		"joinURL": HandleJoinURL,
		"end":     HandleEnd,
	},
	c: map[*Client]struct{}{},
}

func init() {
	http.Handle("/ws", websocket.Server{Handler: HandleWS})
}

func HandleWS(ws *websocket.Conn) {
	remoteAddr := ws.Request().RemoteAddr
	log.Printf("Connection from %s opened", remoteAddr)

	client := &Client{
		address: remoteAddr,
		conn:    ws,
		done:    make(chan struct{}),
		events:  make(chan WsEvent),
	}

	handler.AddClient(client)

	defer func() {
		log.Printf("Connection from %s closed", remoteAddr)
		handler.RemoveClient(client)
	}()

	go client.Writer()
	client.Reader()
}

type Client struct {
	address string
	conn    *websocket.Conn
	b3      bbb.BigBlueButton
	done    chan struct{}
	events  chan WsEvent
	handler *WsEventHandler

	Id string
}

func (c *Client) Reader() {
	for {
		var ev WsEvent
		if err := websocket.JSON.Receive(c.conn, &ev); nil != err {
			if io.EOF == err {
				log.Printf("Reader[%s]: %s", c.address, err)
				c.done <- struct{}{}
				return
			}
		}
		if err := c.handler.Handle(c, ev); nil != err {
			log.Printf("Reader[%s]: %s", c.address, err)
		}
	}
}

func (c *Client) Writer() {
	for {
		select {
		case e := <-c.events:
			log.Printf("Writer[%s]: %#v", c.address, e)
			if err := websocket.JSON.Send(c.conn, e); nil != err {
				log.Printf("Writer[%s]: %s", c.address, err)
			}
		case <-c.done:
			log.Printf("Writer[%s]: exit", c.address)
			return
		}
	}
}

type WsEventData map[string]interface{}

type WsEvent struct {
	Event string      `json:"event"`
	Data  WsEventData `json:"data"`
}

type WsEventHandlerFunc func(*Client, WsEvent) error

type WsEventHandler struct {
	h map[string]WsEventHandlerFunc
	c map[*Client]struct{}
	m sync.RWMutex
}

func (ws *WsEventHandler) Handle(c *Client, ev WsEvent) error {
	if h, t := ws.h[ev.Event]; t {
		return h(c, ev)
	}
	return newWsEventHandlerNotFound(ev.Event)
}

func (ws *WsEventHandler) AddClient(c *Client) {
	ws.m.Lock()
	defer ws.m.Unlock()
	if _, t := ws.c[c]; !t {
		ws.c[c] = struct{}{}
		c.handler = ws
	}
}

func (ws *WsEventHandler) RemoveClient(c *Client) {
	ws.m.Lock()
	defer ws.m.Unlock()
	if _, t := ws.c[c]; t {
		delete(ws.c, c)
		c.handler = nil
	}
}

func (ws *WsEventHandler) Broadcast(event WsEvent) error {
	ws.m.RLock()
	defer ws.m.RUnlock()
	for peer, _ := range ws.c {
		peer.events <- event
	}
	return nil
}

type WsEventHandlerNotFound string

func (e WsEventHandlerNotFound) Error() string {
	return "Event Handler '" + string(e) + "' not found!"
}

func newWsEventHandlerNotFound(e string) WsEventHandlerNotFound {
	return WsEventHandlerNotFound(e)
}
