package main

import "time"

type HTMLdata struct {
	BaseURL   string
	Error     string
	PageTitle string
	ItemCount int    `json:"itemCount"`
	PageCount int    `json:"pageCount"`
	Items     []Path `json:"items"`
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

	// Added internaly
	ID           string `json:"id,omitempty"`
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
