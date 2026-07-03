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
	Name     string  `json:"name"`     // shows the actual url path
	ConfName string  `json:"confName"` // shows the path in the config, so for non-defined path's this shows up as: "all_others"
	Source   Session `json:"source"`

	// v1.19.2 fields (available/online/tracks2 replace the deprecated ready/readyTime/tracks)
	Available            bool        `json:"available"`
	AvailableTime        time.Time   `json:"availableTime"`
	Online               bool        `json:"online"`
	Tracks2              []PathTrack `json:"tracks2"`
	InboundFramesInError uint64      `json:"inboundFramesInError"`
	Readers              []Session   `json:"readers"`

	// Deprecated fields, kept for fallback against older MediaMTX servers
	LegacyReady     bool      `json:"ready"`
	LegacyReadyTime time.Time `json:"readyTime"`
	LegacyTracks    []string  `json:"tracks"`

	// Added internally
	PrettyName   string // internaly created
	PathName     string // internaly created
	Ready        bool   // derived from Available (fallback LegacyReady)
	Tracks       []string
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

// PathTrack is an item of the v1.19.2 tracks2 array.
type PathTrack struct {
	Codec      string         `json:"codec"`
	CodecProps map[string]any `json:"codecProps"`
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
	User               string  `json:"user"`
	UserAgent          string  `json:"userAgent"`
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
	Country             string
	CountryCode         string
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
	User          string `json:"user"`
	UserAgent     string `json:"userAgent"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`
	// Derived
	BytesReceivedStr string
	BytesSentStr     string
	Country          string
	CountryCode      string
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
	Query         string `json:"query"`
	User          string `json:"user"`
	BytesReceived uint64 `json:"bytesReceived"`
	BytesSent     uint64 `json:"bytesSent"`
	// v1.19.2 SRT telemetry
	MsRTT                   float64 `json:"msRTT"`
	MbpsReceiveRate         float64 `json:"mbpsReceiveRate"`
	MbpsSendRate            float64 `json:"mbpsSendRate"`
	MbpsLinkCapacity        float64 `json:"mbpsLinkCapacity"`
	PacketsReceivedLossRate float64 `json:"packetsReceivedLossRate"`
	// Derived
	BytesReceivedStr string
	BytesSentStr     string
	RTTStr           string
	ReceiveRateStr   string
	SendRateStr      string
	LinkCapacityStr  string
	LossRateStr      string
	Country          string
	CountryCode      string
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
	User                      string  `json:"user"`
	UserAgent                 string  `json:"userAgent"`
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
	Country             string
	CountryCode         string
}

type MoQSessionList struct {
	ItemCount int          `json:"itemCount"`
	PageCount int          `json:"pageCount"`
	Items     []MoQSession `json:"items"`
}

// MoQSession represents a Media-over-QUIC session (new in MediaMTX v1.19.0).
type MoQSession struct {
	ID         string `json:"id"`
	Created    string `json:"created"`
	RemoteAddr string `json:"remoteAddr"`
	State      string `json:"state"`
	Path       string `json:"path"`
	Query      string `json:"query"`
	UserAgent  string `json:"userAgent"`
	// Derived
	Country     string
	CountryCode string
}

type HLSSessionList struct {
	ItemCount int          `json:"itemCount"`
	PageCount int          `json:"pageCount"`
	Items     []HLSSession `json:"items"`
}

// HLSSession represents an HLS reader session (new in MediaMTX v1.18.0).
// Unlike HLSMuxer it carries a remoteAddr, so HLS viewers can be geo-located.
type HLSSession struct {
	ID            string `json:"id"`
	Created       string `json:"created"`
	RemoteAddr    string `json:"remoteAddr"`
	Path          string `json:"path"`
	Query         string `json:"query"`
	User          string `json:"user"`
	UserAgent     string `json:"userAgent"`
	IsCDN         bool   `json:"isCDN"`
	OutboundBytes uint64 `json:"outboundBytes"`
	// Derived
	BytesSentStr string
	Country      string
	CountryCode  string
}

// GeoIPResult represents the response from ip-api.com
type GeoIPResult struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Query       string `json:"query"`
}

// CountrySummary shows viewer count per country
type CountrySummary struct {
	Country     string
	CountryCode string
	Count       int
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
	MoQViewers    int
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
	HLSSessionCount   int
	SRTCount          int
	MoQCount          int
	StreamSummaries   []StreamSummary
	CountrySummaries  []CountrySummary
	WebRTCSessions    []WebRTCSession
	RTSPSessions      []RTSPSession
	RTMPConns         []RTMPConn
	HLSMuxers         []HLSMuxer
	HLSSessions       []HLSSession
	SRTConns          []SRTConn
	MoQSessions       []MoQSession
}
