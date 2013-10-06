package bbb

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var EmptyOptions = &emptyOptions{}

type CreateOptions struct {
	Name            string         `json:"name"`
	AttendeePW      string         `json:"attendeePW"`
	ModeratorPW     string         `json:"moderatorPW"`
	Welcome         string         `json:"welcome"`
	DialNumber      string         `json:"dialNumber"`
	VoiceBridge     string         `json:"voiceBridge"`
	WebVoice        string         `json:"webVoice"`
	LogoutURL       string         `json:"logoutURL"`
	MaxParticipants uint           `json:"maxParticipants"`
	Record          bool           `json:"record"`
	Duration        time.Duration  `json:"duration"`
	Documents       []Presentation `json:documents`
}

type JoinOptions struct {
	CreateTime   time.Time `json:"createTime"`
	UserId       string    `json:"userID"`
	WebVoiceConf string    `json:"webVoiceConf"`
}

type Presentation struct {
	Name  string `json:"name,omitempty" xml:"name,attr,omitempty"`
	Url   string `json:"url,omitempty"  xml:"url,attr,omitempty"`
	Value []byte `json:"name,omitempty" xml:",chardata"`
}

type OptionEncoder interface {
	Values() url.Values
}

type emptyOptions struct{}

func (opt *emptyOptions) Values() url.Values { return url.Values{} }

func (opt *CreateOptions) Values() url.Values {
	return reflectOptionValues(reflect.ValueOf(*opt), true,
		func(k string, _ reflect.Value) bool {
			return "documents" != k
		})
}

func (opt *JoinOptions) Values() url.Values {
	return reflectOptionValues(reflect.ValueOf(*opt), true, nil)
}

func reflectOptionValues(rv reflect.Value, skipFalse bool,
	accept func(string, reflect.Value) bool) url.Values {
	values := url.Values{}
	if reflect.Struct == rv.Kind() {
		for i := 0; i < rv.NumField(); i++ {
			if name, value := optionNameFromStructField(rv.Type().Field(i)),
				rv.Field(i); nil == accept || accept(name, value) {
				switch value.Kind() {
				case reflect.Bool:
					if value := value.Bool(); value || !skipFalse {
						values.Set(name, strconv.FormatBool(value))
					}
				case reflect.String:
					if value := value.String(); "" != value {
						values.Set(name, value)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					values.Set(name, strconv.FormatInt(value.Int(), 10))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if value := value.Uint(); value > 0 {
						values.Set(name, strconv.FormatUint(value, 10))
					}
				}
			}
		}
	}
	return values
}

func optionNameFromStructField(s reflect.StructField) string {
	if tag := s.Tag.Get("json"); tag != "" {
		tag, _ := parseTag(tag)
		return tag
	}
	return s.Name
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}
