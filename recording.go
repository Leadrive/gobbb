package bbb

import (
	"time"
)

type Recording struct {
	RecordId  string
	MeetingId string
	Name      string
	Published bool
	StartTime time.Time
	EndTime   time.Time
	Metadata  map[string]interface{}
	Playback  struct {
		Type string
		Url  string
		Len  int
	}
}
