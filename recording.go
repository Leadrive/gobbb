package bbb

import (
	"time"
)

type Recording struct {
	RecordId  string
	MeetingId string
	Name      string
	Published bool
	startTime time.Time
	endTime   time.Time
	Metadata  map[string]interface{}
}
