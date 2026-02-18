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

// Monitoring types

type ServerInfo struct {
	Version string `json:"version"`
	Started string `json:"started"`
	// Derived
	UptimeStr string
}

type HLSMuxerList struct {
	ItemCount int        `json:"itemCount"`
	PageCount int        `json:"pageCount"`
	Items     []HLSMuxer `json:"items"`
}

type HLSMuxer struct {
	Path        string `json:"path"`
	Created     string `json:"created"`
	LastRequest string `json:"lastRequest"`
	BytesSent   uint64 `json:"bytesSent"`
	// Derived
	BytesSentStr   string
	CreatedStr     string
	LastRequestStr string
}

type RTSPSessionList struct {
	ItemCount int           `json:"itemCount"`
	PageCount int           `json:"pageCount"`
	Items     []RTSPSession `json:"items"`
}

type RTSPSession struct {
	ID                 string  `json:"id"`
	Created            string  `json:"created"`
	RemoteAddr         string  `json:"remoteAddr"`
	State              string  `json:"state"`
	Path               string  `json:"path"`
	Query              string  `json:"query"`
	Transport          string  `json:"transport"`
	BytesReceived      uint64  `json:"bytesReceived"`
	BytesSent          uint64  `json:"bytesSent"`
	RTPPacketsReceived uint64  `json:"rtpPacketsReceived"`
	RTPPacketsSent     uint64  `json:"rtpPacketsSent"`
	RTPPacketsLost     uint64  `json:"rtpPacketsLost"`
	RTPPacketsInError  uint64  `json:"rtpPacketsInError"`
	RTPPacketsJitter   float64 `json:"rtpPacketsJitter"`
	// Derived
	BytesReceivedStr    string
	BytesSentStr        string
	RTPPacketsJitterStr string
}

type RTMPConnList struct {
	ItemCount int        `json:"itemCount"`
	PageCount int        `json:"pageCount"`
	Items     []RTMPConn `json:"items"`
}

type RTMPConn struct {
	ID            string `json:"id"`
	Created       string `json:"created"`
	RemoteAddr    string `json:"remoteAddr"`
	State         string `json:"state"`
	Path          string `json:"path"`
	Query         string `json:"query"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`
	// Derived
	BytesReceivedStr string
	BytesSentStr     string
}

type SRTConnList struct {
	ItemCount int       `json:"itemCount"`
	PageCount int       `json:"pageCount"`
	Items     []SRTConn `json:"items"`
}

type SRTConn struct {
	ID            string `json:"id"`
	Created       string `json:"created"`
	RemoteAddr    string `json:"remoteAddr"`
	State         string `json:"state"`
	Path          string `json:"path"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`
	// Derived
	BytesReceivedStr string
	BytesSentStr     string
}

type WebRTCSessionList struct {
	ItemCount int             `json:"itemCount"`
	PageCount int             `json:"pageCount"`
	Items     []WebRTCSession `json:"items"`
}

type WebRTCSession struct {
	ID                        string  `json:"id"`
	Created                   string  `json:"created"`
	RemoteAddr                string  `json:"remoteAddr"`
	PeerConnectionEstablished bool    `json:"peerConnectionEstablished"`
	LocalCandidate            string  `json:"localCandidate"`
	RemoteCandidate           string  `json:"remoteCandidate"`
	State                     string  `json:"state"`
	Path                      string  `json:"path"`
	Query                     string  `json:"query"`
	BytesReceived             uint64  `json:"bytesReceived"`
	BytesSent                 uint64  `json:"bytesSent"`
	RTPPacketsReceived        uint64  `json:"rtpPacketsReceived"`
	RTPPacketsSent            uint64  `json:"rtpPacketsSent"`
	RTPPacketsLost            uint64  `json:"rtpPacketsLost"`
	RTPPacketsJitter          float64 `json:"rtpPacketsJitter"`
	// Derived
	BytesReceivedStr    string
	BytesSentStr        string
	RTPPacketsJitterStr string
}

type StreamSummary struct {
	Name          string
	SourceType    string
	Tracks        []string
	WebRTCViewers int
	RTSPViewers   int
	RTMPViewers   int
	HLSViewers    int
	SRTViewers    int
	TotalViewers  int
	BandwidthStr  string
	TotalBytes    uint64
}

type MonitoringHTMLdata struct {
	BaseURL           string
	Error             string
	PageTitle         string
	Version           string
	ServerInfo        ServerInfo
	TotalStreams      int
	TotalViewers      int
	TotalBandwidthStr string
	WebRTCCount       int
	RTSPCount         int
	RTMPCount         int
	HLSCount          int
	SRTCount          int
	StreamSummaries   []StreamSummary
	WebRTCSessions    []WebRTCSession
	RTSPSessions      []RTSPSession
	RTMPConns         []RTMPConn
	HLSMuxers         []HLSMuxer
	SRTConns          []SRTConn
}
