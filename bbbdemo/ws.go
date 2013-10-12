package main

import (
	"encoding/json"
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
	var options bbb.CreateOptions
	eventToOptions(event, &options)

	if m, err := c.b3.Create(id, &options); nil != err {
		ev := WsEvent{"create.fail", WsEventData{"error": err.Error()}}
		if v, t := event.Data["__txid"]; t {
			ev.Data["__txid"] = v.(string)
		}
		c.events <- ev
	} else {
		ev := WsEvent{"create.success", WsEventData{
			"id":          m.Id,
			"created":     m.CreateTime.Unix(),
			"attendeePW":  m.AttendeePW,
			"moderatorPW": m.ModeratorPW,
			"forcedEnd":   m.ForcedEnd,
		}}
		if v, t := event.Data["__txid"]; t {
			ev.Data["__txid"] = v.(string)
		}
		c.handler.Broadcast(ev)
	}
	return nil
}

func HandleJoinURL(c *Client, event WsEvent) error {
	name, id, password := "", "", ""
	if v, t := event.Data["name"]; t && nil != v {
		name = v.(string)
	}
	if v, t := event.Data["fullName"]; t && nil != v {
		name = v.(string)
	}
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}
	if v, t := event.Data["password"]; t && nil != v {
		password = v.(string)
	}
	var options bbb.JoinOptions
	eventToOptions(event, &options)
	c.events <- WsEvent{"joinURL", WsEventData{
		"url": c.b3.JoinURL(name, id, password, &options),
	}}
	return nil
}

func HandleEnd(c *Client, event WsEvent) error {
	id, password := "", ""
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}
	if v, t := event.Data["password"]; t && nil != v {
		password = v.(string)
	}
	ev := WsEvent{"end", WsEventData{"ended": false, "id": id}}
	if v, t := event.Data["__txid"]; t {
		ev.Data["__txid"] = v.(string)
	}
	if ok := b3.End(id, password); ok {
		ev.Data["ended"] = true
		c.handler.Broadcast(ev)
	} else {
		c.events <- ev
	}
	return nil
}

func HandleIsMeetingRunning(c *Client, event WsEvent) error {
	id := ""
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}
	c.events <- WsEvent{"running", WsEventData{
		"running": c.b3.IsMeetingRunning(id)},
	}
	return nil
}

func HandleMeetingInfo(c *Client, event WsEvent) error {
	id, password := "", ""
	if v, t := event.Data["id"]; t && nil != v {
		id = v.(string)
	}
	if v, t := event.Data["password"]; t && nil != v {
		password = v.(string)
	}
	m, err := c.b3.MeetingInfo(id, password)
	if nil != err {
		c.events <- WsEvent{"info.fail", WsEventData{"error": err.Error()}}
		return nil
	}
	attendees := make([]WsEventData, len(m.Attendees))
	for k, v := range m.Attendees {
		attendees[k] = WsEventData{
			"userID": v.UserId,
			"name":   v.Name,
			"role":   v.Role,
		}
	}
	c.events <- WsEvent{"info.succsess", WsEventData{
		"id":          m.Id,
		"name":        m.Name,
		"created":     m.CreateTime.Unix(),
		"attendeePW":  m.AttendeePW,
		"moderatorPW": m.ModeratorPW,
		"running":     m.Running,
		"recording":   m.Recording,
		"forcedEnd":   m.ForcedEnd,
		"startTime":   m.StartTime.Unix(),
		"endTime":     m.EndTime.Unix(),
		"numUsers":    m.NumUsers,
		"maxUsers":    m.MaxUsers,
		"numMod":      m.NumMod,
		"attendees":   attendees,
	}}
	return nil
}

func HandleMeetings(c *Client, event WsEvent) error {
	meetings := c.b3.Meetings()
	ev := make([]WsEventData, len(meetings))
	for k, m := range meetings {
		ev[k] = WsEventData{
			"id":          m.Id,
			"created":     m.CreateTime.Unix(),
			"attendeePW":  m.AttendeePW,
			"moderatorPW": m.ModeratorPW,
			"forcedEnd":   m.ForcedEnd,
		}
	}
	c.events <- WsEvent{"meetings", WsEventData{"meetings": ev}}
	return nil
}

func HandleRecordings(c *Client, event WsEvent) (err error) {
	var meetings []string
	if v, t := event.Data["meetings"]; t {
		meetings = itos(v)
	}
	recordings := c.b3.Recordings(meetings)
	ev := make([]WsEventData, len(recordings))
	for k, r := range recordings {
		ev[k] = WsEventData{
			"recordId":  r.RecordId,
			"meetingId": r.MeetingId,
			"name":      r.Name,
			"startTime": r.StartTime.Unix(),
			"endTime":   r.EndTime.Unix(),
			"playback": WsEventData{
				"type": r.Playback.Type,
				"url":  r.Playback.Url,
				"len":  r.Playback.Len,
			},
		}
	}
	c.events <- WsEvent{"recordings", WsEventData{"recordings": ev}}
	return
}

func HandlePublishRecordings(c *Client, event WsEvent) error {
	recordings, publish := []string{}, true
	if v, t := event.Data["recordings"]; t {
		recordings = itos(v)
	}
	if v, t := event.Data["publish"]; t {
		publish = v.(bool)
	}
	ev := WsEvent{"recordings", WsEventData{
		"recordings": recordings,
		"published":  false,
	}}
	if v, t := event.Data["__txid"]; t {
		ev.Data["__txid"] = v.(string)
	}
	if c.b3.PublishRecordings(recordings, publish) {
		ev.Data["published"] = true
		c.handler.Broadcast(ev)
	} else {
		c.events <- ev
	}
	return nil
}

func HandleDeleteRecordings(c *Client, event WsEvent) error {
	var recordings []string
	if v, t := event.Data["recordings"]; t {
		recordings = itos(v)
	}
	ev := WsEvent{"recordings", WsEventData{
		"recordings": recordings,
		"deleted":    false,
	}}
	if v, t := event.Data["__txid"]; t {
		ev.Data["__txid"] = v.(string)
	}
	if c.b3.DeleteRecordings(recordings) {
		ev.Data["deleted"] = true
		c.handler.Broadcast(ev)
	} else {
		c.events <- ev
	}
	return nil
}

func HandleDefaultConfigXML(c *Client, event WsEvent) error {
	if conf, err := c.b3.DefaultConfigXML(); nil != err {
		c.events <- WsEvent{"config.error", WsEventData{
			"error": err.Error(),
		}}
	} else {
		var data WsEventData
		if err := jsoncp(&data, conf); nil != err {
			return err
		}
		c.events <- WsEvent{"config.default", data}
	}
	return nil
}

func HandleSetConfigXML(c *Client, event WsEvent) error {
	var conf bbb.ConfigXML
	var config interface{}
	var meeting = ""
	if v, t := event.Data["config"]; t {
		config = v
	}
	if v, t := event.Data["meeting"]; t {
		meeting = v.(string)
	}
	if err := jsoncp(&conf, config); nil != err {
		return err
	}
	if token, err := c.b3.SetConfigXML(meeting, &conf); nil != err {
		c.events <- WsEvent{"config.error", WsEventData{
			"error": err.Error(),
		}}
	} else {
		c.events <- WsEvent{"config.set", WsEventData{
			"meeting": meeting,
			"token":   token,
		}}
	}
	return nil
}

var handler *WsEventHandler = &WsEventHandler{
	h: map[string]WsEventHandlerFunc{
		"connect":  HandleConnect,
		"create":   HandleCreate,
		"joinURL":  HandleJoinURL,
		"end":      HandleEnd,
		"running":  HandleIsMeetingRunning,
		"info":     HandleMeetingInfo,
		"meetings": HandleMeetings,

		"recordings":         HandleRecordings,
		"recordings.publish": HandlePublishRecordings,
		"recordings.delete":  HandleDeleteRecordings,

		"config.default": HandleDefaultConfigXML,
		"config.set":     HandleSetConfigXML,
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

func eventToOptions(event WsEvent, options interface{}) error {
	if b, err := json.Marshal(event.Data); nil == err {
		return json.Unmarshal(b, options)
	} else {
		return err
	}
}

func itos(v interface{}) (s []string) {
	if v, ok := v.([]interface{}); ok {
		s = make([]string, len(v))
		for k, v := range v {
			s[k] = v.(string)
		}
		return

	}
	panic("Too bad")
}

func jsoncp(dst interface{}, src interface{}) error {
	if b, err := json.Marshal(src); nil == err {
		if err := json.Unmarshal(b, dst); nil != err {
			return err
		}
	} else {
		return err
	}
	return nil
}
