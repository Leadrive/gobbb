package bbb

import (
	"net/url"
	"strconv"
	"time"
)

type Options interface {
	Encode() string
	Decode(string)
	T() OptionType
}

type OptionType int

const (
	OptionsEmpty OptionType = iota
	OptionsCreate
	OptionsJoin
)

var EmptyOptions = &xOptions{make(map[string]interface{}), OptionsEmpty}

var optionKeys [][]string = [][]string{
	{},
	{"name", "attendeePW", "moderatorPW", "welcome", "dialNumber", "voiceBridge",
		"webVoice", "logoutURL", "maxParticipants", "record", "duration"},
	{"createTime", "userID", "webVoiceConf"},
}

type xOptions struct {
	values map[string]interface{}
	typ    OptionType
}

func (opt *xOptions) Encode() string {
	var values url.Values
	var str string
	for k, v := range opt.values {
		if isOptionKeyAllowed(k, opt) {
			switch v := v.(type) {
			case bool:
				str = strconv.FormatBool(v)
			case []byte:
				str = string(v)
			case float32, float64:
				str = strconv.FormatFloat(v.(float64), 'f', -1, 32)
			case int, int8, int16, int32, int64, time.Duration:
				str = strconv.FormatInt(v.(int64), 10)
			case uint, uint8, uint16, uint32, uint64:
				str = strconv.FormatUint(v.(uint64), 10)
			case string:
				str = v
			// case time.Duration:
			// 	str = strconv.FormatInt(int64(v), 10)
			case time.Time:
				str = strconv.FormatInt(v.Unix(), 10)
			case nil:
				continue
			default:
				panic("unhandled type")
			}
			values.Set(k, str)
		}
	}
	return values.Encode()
}

func (opt *xOptions) Decode(_ string) {
	panic("not implemented")
}

func (opt *xOptions) T() OptionType {
	return opt.typ
}

func isOptionKeyAllowed(k string, o Options) bool {
	if t := o.T(); t < OptionType(len(optionKeys)) {
		for _, v := range optionKeys[t] {
			if v == k {
				return true
			}
		}
	}
	return false
}
