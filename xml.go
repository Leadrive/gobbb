package bbb

import (
	"net/http"
	"time"

	"github.com/sdgoij/go-pkg-xmlx"
)

func loadResponseXML(r *http.Response) (response *xmlx.Node, err error) {
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

func loadMeetingCreateResponse(r *http.Response) (*Meeting, error) {
	if response, err := loadResponseXML(r); nil == err {
		return xml2meeting(response), nil
	} else {
		return nil, err
	}
}

func loadMeetingInfoResponse(r *http.Response) (*Meeting, error) {
	if response, err := loadResponseXML(r); nil == err {
		nodes := response.SelectNodes("", "attendee")
		attendees := make([]Attendee, len(nodes))
		for k, v := range nodes {
			attendees[k] = Attendee{
				UserId: v.S("", "userID"),
				Name:   v.S("", "fullName"),
				Role:   v.S("", "role"),
			}
		}
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
			Attendees:   attendees,
		}, nil
	} else {
		return nil, err
	}
}

func loadMeetigsResponse(r *http.Response) []*Meeting {
	if response, err := loadResponseXML(r); nil == err {
		if nodes := response.SelectNode("", "meetings").SelectNodes("", "meeting"); len(nodes) > 0 {
			meetings := make([]*Meeting, len(nodes))
			for index, meeting := range nodes {
				meetings[index] = xml2meeting(meeting)
			}
			return meetings
		}
	}
	return []*Meeting{}
}

func loadRecordingsResponse(r *http.Response) []*Recording {
	if response, err := loadResponseXML(r); nil == err {
		if nodes := response.SelectNode("", "recordings").SelectNodes("", "recording"); len(nodes) > 0 {
			recordings := make([]*Recording, len(nodes))
			for index, recording := range nodes {
				playback := recording.SelectNode("", "playback")
				recordings[index] = &Recording{
					RecordId:  recording.S("", "recordId"),
					MeetingId: recording.S("", "meetingId"),
					Name:      recording.S("", "name"),
					StartTime: time.Unix(recording.I64("", "startTime"), 0),
					EndTime:   time.Unix(recording.I64("", "endTime"), 0),
					Playback: struct {
						Type string
						Url  string
						Len  int
					}{
						playback.S("", "type"),
						playback.S("", "url"),
						playback.I("", "length"),
					},
				}
			}
			return recordings
		}
	}
	return []*Recording{}
}

func loadBoolResponse(r *http.Response, element string) bool {
	if response, err := loadResponseXML(r); nil == err {
		return response.B("", element)
	}
	return false
}

func loadStringResponse(r *http.Response, element string) string {
	if response, err := loadResponseXML(r); nil == err {
		return response.S("", element)
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

func buildCreateMeetingXML(docs []ConfigXML_Document) ([]byte, error) {
	doc, config := xmlx.New(), &ConfigXML{
		Modules: []ConfigXML_Module{
			{Name: "presentation", Documents: docs},
		},
	}
	if err := doc.LoadString(config.String(), nil); nil != err {
		return []byte(nil), err
	}
	return []byte(doc.SelectNode("", "modules").String()), nil
}

func mstime(ts int64) time.Time {
	return time.Unix(int64(ts/int64(time.Microsecond)), 0)
}

type xmlError string

func (err xmlError) Error() string {
	return string(err)
}
