package bbb

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func New(apiurl, secret string) (BigBlueButton, error) {
	b3 := BigBlueButton{Secret: secret}
	u, err := url.Parse(apiurl)
	if nil == err {
		b3.Url = u
	}
	return b3, err
}

type BigBlueButton struct {
	Secret string
	Url    *url.URL
}

func (b3 *BigBlueButton) Create(id string, options Options) (*Meeting, error) {
	v := url.Values{"meetingID": {id}}
	q := v.Encode() // + "&" + options.Encode()
	q += "&checksum=" + b3.checksum("create", q)
	u, _ := b3.Url.Parse("create?" + q)
	log.Println(u.String())
	res, err := http.Get(u.String())
	if nil != err {
		return nil, err
	}
	defer res.Body.Close()
	return LoadMeetingCreateResponse(res)
}

func (b3 *BigBlueButton) JoinURL(name, meetingID, password string, options Options) string {
	v := url.Values{}
	v.Set("fullName", name)
	v.Set("meetingID", meetingID)
	v.Set("password", password)
	q := v.Encode() // + "&" + options.Encode()
	q += "&checksum=" + b3.checksum("join", q)
	u, _ := b3.Url.Parse("join?" + q)
	return u.String()
}

func (b3 *BigBlueButton) IsMeetingRunning(id string) bool {
	v := url.Values{}
	v.Set("meetingID", id)
	q := v.Encode()
	q += "&checksum=" + b3.checksum("isMeetingRunning", q)
	u, _ := b3.Url.Parse("isMeetingRunning?" + q)
	res, err := http.Get(u.String())
	if nil != err {
		return false
	}
	defer res.Body.Close()
	return LoadIsMeetingRunningResponse(res)
}

func (b3 *BigBlueButton) End(id, password string) bool {
	time.Sleep(2 * time.Second)
	return b3.IsMeetingRunning(id)
}

func (b3 *BigBlueButton) MeetingInfo(id, password string) (*Meeting, error) {
	v := url.Values{"meetingID": {id}, "password": {password}}
	q := v.Encode()
	q += "&checksum=" + b3.checksum("getMeetingInfo", q)
	u, _ := b3.Url.Parse("getMeetingInfo?" + q)
	res, err := http.Get(u.String())
	if nil != err {
		return nil, err
	}
	defer res.Body.Close()
	return LoadMeetingInfoResponse(res)
}

func (b3 *BigBlueButton) Meetings() []*Meeting {
	q := "checksum=" + b3.checksum("getMeetings", "")
	u, _ := b3.Url.Parse("getMeetings?" + q)
	log.Println(u.String())
	res, err := http.Get(u.String())
	if nil != err {
		return []*Meeting{}
	}
	defer res.Body.Close()
	return LoadMeetigsResponse(res)
}

func (b3 *BigBlueButton) Recordings(meetingId string) []*Recording {
	return []*Recording(nil)
}

func (b3 *BigBlueButton) PublishRecordings(id string, publish bool) bool {
	return false
}

func (b3 *BigBlueButton) DeleteRecordings(id string) bool {
	return false
}

func (b3 *BigBlueButton) ServerVersion() string {
	res, err := http.Get(b3.Url.String())
	if nil != err {
		return "Error: " + err.Error()
	}
	defer res.Body.Close()
	return LoadServerVersion(res)
}

func (b3 *BigBlueButton) checksum(action, params string) string {
	if i := len(params) - 1; i > 0 && params[i] == '&' {
		params = params[:i]
	}
	h := sha1.New()
	io.WriteString(h, action)
	io.WriteString(h, params)
	io.WriteString(h, b3.Secret)
	return fmt.Sprintf("%x", h.Sum(nil))
}
