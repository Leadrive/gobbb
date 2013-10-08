package bbb

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/sdgoij/go-pkg-xmlx"
)

func LoadResponseXML(r *http.Response) (response *xmlx.Node, err error) {
	var doc *xmlx.Document = xmlx.New()
	if err = doc.LoadStream(r.Body, nil); nil != err {
		return
	}
	response = doc.SelectNode("", "response")
	if code := response.S("", "returncode"); code != "SUCCESS" {
		err = xmlError(code + " " + response.S("", "messageKey"))
	}
	return
}

func LoadMeetingCreateResponse(r *http.Response) (*Meeting, error) {
	if response, err := LoadResponseXML(r); nil == err {
		return xml2meeting(response), nil
	} else {
		return nil, err
	}
}

func LoadMeetingInfoResponse(r *http.Response) (*Meeting, error) {
	if response, err := LoadResponseXML(r); nil == err {
		return &Meeting{
			Id:          response.S("", "meetingID"),
			Name:        response.S("", "meetingName"),
			CreateTime:  mstime(response.I64("", "createTime")),
			VoiceBridge: response.I("", "voiceBridge"),
			AttendeePW:  response.S("", "attendeePW"),
			ModeratorPW: response.S("", "moderatorPW"),
			Running:     response.B("", "running"),
			Recording:   response.B("", "recording"),
			ForcedEnd:   response.B("", "hasBeenForciblyEnded"),
			StartTime:   mstime(response.I64("", "startTime")),
			EndTime:     mstime(response.I64("", "endTime")),
			NumUsers:    response.I("", "participantCount"),
			NumMod:      response.I("", "moderatorCount"),
			MaxUsers:    response.I("", "maxUsers"),
		}, nil
	} else {
		return nil, err
	}
}

func LoadIsMeetingRunningResponse(r *http.Response) bool {
	if response, err := LoadResponseXML(r); nil == err {
		return response.B("", "running")
	}
	return false
}

func LoadMeetigsResponse(r *http.Response) []*Meeting {
	if response, err := LoadResponseXML(r); nil == err {
		if nodes := response.SelectNodes("", "meeting"); len(nodes) > 0 {
			meetings := make([]*Meeting, len(nodes))
			for index, meeting := range nodes {
				meetings[index] = xml2meeting(meeting)
			}
			return meetings
		}
	}
	return []*Meeting{}
}

func LoadServerVersion(r *http.Response) string {
	if response, err := LoadResponseXML(r); nil == err {
		return response.S("", "version")
	} else {
		return err.Error()
	}
}

func xml2meeting(meeting *xmlx.Node) *Meeting {
	if nil != meeting {
		return &Meeting{
			Id:          meeting.S("", "meetingID"),
			Name:        meeting.S("", "meetingName"),
			CreateTime:  mstime(meeting.I64("", "createTime")),
			AttendeePW:  meeting.S("", "attendeePW"),
			ModeratorPW: meeting.S("", "moderatorPW"),
			ForcedEnd:   meeting.B("", "hasBeenForciblyEnded"),
		}
	}
	return &Meeting{}
}

type ModXML_Module struct {
	Name string         `xml:"name,attr"`
	Docs []Presentation `xml:"document"`
}

type ModXML_Root struct {
	XMLName xml.Name        `xml:"modules"`
	Modules []ModXML_Module `xml:"module"`
}

func buildModXML_Presentation(docs []Presentation) ([]byte, error) {
	return xml.Marshal(ModXML_Root{xml.Name{}, []ModXML_Module{{"presentation", docs}}})
}

func mstime(ts int64) time.Time {
	return time.Unix(int64(ts/int64(time.Microsecond)), 0)
}

type xmlError string

func (err xmlError) Error() string {
	return string(err)
}
