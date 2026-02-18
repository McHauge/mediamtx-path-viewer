package main

import "time"

type HTMLdata struct {
	BaseURL   string
	Error     string
	PageTitle string
	ItemCount int    `json:"itemCount"`
	PageCount int    `json:"pageCount"`
	Items     []Path `json:"items"`
	Groups    []PathGroup
	Version   string
}

type PathGroup struct {
	GroupName string
	Paths     []Path
}

type MediaMTX struct {
	ItemCount int    `json:"itemCount"`
	PageCount int    `json:"pageCount"`
	Items     []Path `json:"items"`
}

type Path struct {
	Name       string    `json:"name"`     // shows the actual url path
	ConfName   string    `json:"confName"` // shows the path in the config, so for non-defined path's this shows up as: "all_others"
	PrettyName string    // internaly created
	PathName   string    // internaly created
	Source     Session   `json:"source"`
	Ready      bool      `json:"ready"`
	ReadyTime  time.Time `json:"readyTime"`
	Tracks     []string  `json:"tracks"`
	Readers    []Session `json:"readers"`

	// Added internally
	ID           string `json:"id,omitempty"`
	IsRecording  bool
	ReadyTimeStr string `json:"readyTimeStr,omitempty"`
	TotalReaders int    `json:"totalReaders,omitempty"`
	StreamUrl    string `json:"streamUrl,omitempty"`
	StreamHls    string `json:"StreamHls,omitempty"`
	StreamWebRTC string `json:"StreamWebRTC,omitempty"`
	StreamRtmp   string `json:"StreamRtmp,omitempty"`
	StreamRtsp   string `json:"StreamRtsp,omitempty"`
}

type Session struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// Recording types

type RecordingList struct {
	ItemCount int         `json:"itemCount"`
	PageCount int         `json:"pageCount"`
	Items     []Recording `json:"items"`
}

type Recording struct {
	Name     string    `json:"name"`
	Segments []Segment `json:"segments"`
	// Derived fields
	ID            string
	PrettyName    string
	PathName      string
	SegmentCount  int
	OldestSegment string
	NewestSegment string
}

type Segment struct {
	Start time.Time `json:"start"`
}

type PlaybackSegment struct {
	Start    time.Time `json:"start"`
	Duration float64   `json:"duration"`
	Url      string    `json:"url"`
	// Derived
	StartStr    string
	DurationStr string
	StartRFC    string
}

type RecordingGroup struct {
	GroupName  string
	Recordings []Recording
}

type RecordingsHTMLdata struct {
	BaseURL     string
	Error       string
	PageTitle   string
	ItemCount   int
	PageCount   int
	Items       []Recording
	Groups      []RecordingGroup
	Version     string
	PlaybackURL string
}

type RecordingDetailData struct {
	Recording
	BaseURL     string
	PlaybackURL string
	Segments    []PlaybackSegment
}
