package bbb

import (
	"encoding/xml"
	"io"
	"io/ioutil"
)

// More information about ConfigXML (config.xml) see:
//   https://code.google.com/p/bigbluebutton/wiki/ClientConfiguration

type ConfigXML struct {
	Version       string                  `json:"version" xml:"version"`
	LocaleVersion ConfigXML_LocaleVersion `json:"localeversion" xml:"localeversion"`
	Help          ConfigXML_Help          `json:"help,omitempty" xml:"help,omitempty"`
	BwMon         ConfigXML_BwMon         `json:"bwMon,omitempty" xml:"bwMon,omitempty"`
	Application   ConfigXML_Application   `json:"application,omitempty" xml:"application,omitempty"`
	Language      ConfigXML_Language      `json:"language,omitempty" xml:"language,omitempty"`
	Layout        ConfigXML_Layout        `json:"layout,omitempty" xml:"layout,omitempty"`
	Modules       []ConfigXML_Module      `json:"modules,omitempty" xml:"modules>module,omitempty"`
}

func (c *ConfigXML) String() string {
	if x, err := xml.Marshal(c); nil == err {
		return string(x)
	}
	return ""
}

func readConfigXML(r io.Reader) (c *ConfigXML, e error) {
	var conf ConfigXML
	if data, err := ioutil.ReadAll(r); nil == err {
		e = xml.Unmarshal(data, &conf)
		c = &conf
	} else {
		e = err
	}
	return
}

type ConfigXML_Application struct {
	Uri  string `json:"uri,omitempty" xml:"uri,attr,omitempty"`
	Host string `jsin:"host,omitempty" xml:"host,attr,omitempty"`
}

type ConfigXML_BwMon struct {
	Server      string `json:"server,omitempty" xml:"server,attr,omitempty"`
	Application string `json:"application,omitempty" xml:"application,attr,omitempty"`
}

type ConfigXML_Document struct {
	Name  string `json:"name,omitempty" xml:"name,attr,omitempty"`
	Url   string `json:"url,omitempty"  xml:"url,attr,omitempty"`
	Value []byte `json:"name,omitempty" xml:",chardata"`
}

type ConfigXML_Help struct {
	Url string `json:"url,omitempty" xml:"url,attr,omitempty"`
}

type ConfigXML_Language struct {
	UserSelectionEnabled bool `json:"userSelectionEnabled" xml:"userSelectionEnabled,attr"`
}

type ConfigXML_Layout struct {
	DefaultLayout      string `json:"defaultLayout,omitempty" xml:"defaultLayout,attr,omitempty"`
	ShowLogButton      bool   `json:"showLogButton" xml:"showLogButton,attr"`
	ShowVideoLayout    bool   `json:"showVideoLayout" xml:"showVideoLayout,attr"`
	ShowResetLayout    bool   `json:"showResetLayout" xml:"showResetLayout,attr"`
	ShowToolbar        bool   `json:"showToolbar" xml:"showToolbar,attr"`
	ShowFooter         bool   `json:"showFooter" xml:"showFooter,attr"`
	ShowMeetingName    bool   `json:"showMeetingName" xml:"showMeetingName,attr"`
	ShowHelpButton     bool   `json:"showHelpButton" xml:"showHelpButton,attr"`
	ShowLogoutWindow   bool   `json:"showLogoutWindow" xml:"showLogoutWindow,attr"`
	ShowLayoutTools    bool   `json:"showLayoutTools" xml:"showLayoutTools,attr"`
	ShowNetworkMonitor bool   `json:"showNetworkMonitor" xml:"showNetworkMonitor,attr"`
	ConfirmLogout      bool   `json:"confirmLogout" xml:"confirmLogout,attr"`
}

type ConfigXML_LocaleVersion struct {
	SuppressWarning bool   `json:"suppressWarning" xml:"suppressWarning,attr"`
	Version         string `json:"version" xml:",chardata"`
}

type ConfigXML_Module struct {
	Name         string `json:"name" xml:"name,attr"`
	Url          string `json:"url" xml:"url,attr"`
	Uri          string `json:"uri" xml:"uri,attr"`
	BaseTabIndex int    `json:"baseTabIndex,omitempty" xml:"baseTabIndex,attr,omitempty"`
	DependsOn    string `json:"dependsOn,omitempty" xml:"dependsOn,attr,omitempty"`
	Position     string `json:"position,omitempty" xml:"position,attr,omitempty"`
	ShowButton   bool   `json:"showButton,omitempty" xml:"showButton,attr,omitempty"`

	// Layout Module
	LayoutConfig string `json:"layoutConfig,omitempty" xml:"layoutConfig,attr,omitempty"`
	EnableEdit   bool   `json:"enableEdit,omitempty" xml:"enableEdit,attr,omitempty"`

	// Chat Module
	TranslationOn      bool `json:"translationOn,omitempty" xml:"translationOn,attr,omitempty"`
	TranslationEnabled bool `json:"translationEnabled,omitempty" xml:"translationEnabled,attr,omitempty"`
	PrivateEnabled     bool `json:"privateEnabled,omitempty" xml:"privateEnabled,attr,omitempty"`

	// Viewers Module
	AllowKick bool `json:"allowKickUser,omitempty" xml:"allowKickUser,attr,omitempty"`
	Visible   bool `json:"windowVisible,omitempty" xml:"windowVisible,attr,omitempty"`

	// Desktop Sharing, Phone Module
	AutoStart  bool `json:"autoStart,omitempty" xml:"autoStart,attr,omitempty"`
	AutoJoin   bool `json:"autoJoin,omitempty" xml:"autoJoin,attr,omitempty"`
	CancelEcho bool `json:"enabledEchoCancel,omitempty" xml:"enabledEchoCancel,attr,omitempty"`

	// Videoconf Module
	PresenterShareOnly   bool   `json:"presenterShareOnly,omitempty" xml:"presenterShareOnly,attr,omitempty"`
	Resolutions          string `json:"resolutions,omitempty" xml:"resolutions,attr,omitempty"`
	PublishWindow        bool   `json:"publishWindowVisible,omitempty" xml:"publishWindowVisible,attr,omitempty"`
	ViewerWindowMaxed    bool   `json:"viewerWindowMaxed,omitempty" xml:"viewerWindowMaxed,attr,omitempty"`
	ViewerWindowLocation string `json:"viewerWindowLocation,omitempty" xml:"viewerWindowLocation,attr,omitempty"`
	CamKeyFrameInterval  int    `json:"camKeyFrameInterval,omitempty" xml:"camKeyFrameInterval,attr,omitempty"`
	CamModeFps           int    `json:"camModeFps,omitempty" xml:"camModeFps,attr,omitempty"`
	CamQualityBandwidth  int    `json:"camQualityBandwidth,omitempty" xml:"camQualityBandwidth,attr,omitempty"`
	CamQualityPicture    int    `json:"camQualityPicture,omitempty" xml:"camQualityPicture,attr,omitempty"`
	H264Level            string `json:"h264Level,omitempty" xml:"h264Level,attr,omitempty"`
	H264Profile          string `json:"h264Profile,attr,omitempty" xml:"h264Profile,attr,omitempty"`

	// Videodock Module
	AutoDock        bool `json:"autoDock,omitempty" xml:"autoDock,attr,omitempty"`
	MaximizeWindow  bool `json:"maximizeWindow,omitempty" xml:"maximizeWindow,attr,omitempty"`
	OneAlwaysBigger bool `json:"oneAlwaysBigger,attr,omitempty" xml:"oneAlwaysBigger,attr,omitempty"`

	// Present Module
	ShowPresentWindow  bool `json:"showPresentWindow,omitempty" xml:"showPresentWindow,attr,omitempty"`
	ShowWindowControls bool `json:"showWindowControls,omitempty" xml:"showWindowControls,attr,omitempty"`

	// Dynamic Info Module (experimental)
	InfoURL string `json:"infoURL,omitempty" xml:"infoURL,attr,omitempty"`

	// Example Chat Module,Breakout Module
	Host   string `json:"host,omitempty" xml:"host,attr,omitempty"`
	Secret string `json:"salt,omitempty" xml:"salt,attr,omitempty"`

	Documents []ConfigXML_Document `json:"documents,omitempty" xml:"document,omitempty"`
}

type ConfigXML_PortTest struct {
	Host        string `json:"host,omitempty" xml:"host,attr,omitempty"`
	Application string `json:"application,omitempty" xml:"application,attr,omitempty"`
	Timeout     int    `json:"timeout,omitempty" xml:"timeout,attr,omitempty"`
}

type ConfigXML_ShortcutKeys struct {
	ShowButton bool `json:"showButton,omitempty" xml:"showButton,attr,omitempty"`
}

type ConfigXML_Skinning struct {
	Enabled bool   `json:"enabled,omitempty" xml:"enabled,attr,omitempty"`
	Url     string `json:"url,omitempty" xml:"url,attr,omitempty"`
}
