package bbb

import (
	"time"
)

type Meeting struct {
	Id          string
	Name        string
	CreateTime  time.Time
	VoiceBridge int
	AttendeePW  string
	ModeratorPW string
	Running     bool
	Recording   bool
	ForcedEnd   bool
	StartTime   time.Time
	EndTime     time.Time
	NumUsers    int
	NumMod      int
	MaxUsers    int
	Attendees   []Attendee
}

type Attendee struct {
	UserId string
	Name   string
	Role   string
}
