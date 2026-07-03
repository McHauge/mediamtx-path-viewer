package main

import (
	"bytes"
	"crypto/tls"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"mediamtx-path-viewer/version"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	log "github.com/s00500/env_logger"
)

//go:generate sh injectGitVars.sh

const (
	appName         = "mediamtx-path-viewer"
	appFriendlyName = "MediaMTX Path Viewer"
	maker           = "McHauge"
)

// HTMLdata is the data structure for the HTML templates
var (
	//go:embed static
	res embed.FS
)

var (
	// Environment Variables
	MEDIAMTX_API_URL  string
	MEDIAMTX_API_PORT string
	MEDIAMTX_USERNAME string
	MEDIAMTX_PASSWORD string

	MEDIAMTX_WEBRTC_URL   string
	MEDIAMTX_HLS_URL      string
	MEDIAMTX_RTMP_URL     string
	MEDIAMTX_RTSP_URL     string
	MEDIAMTX_MOQ_URL      string
	MEDIAMTX_PLAYBACK_URL string

	// Default values
	basePath = ""     // Default to /monitor
	port     = "8080" // Default to 8080

	// Boot Params
	// Version flag
	Version = flag.Bool("version", false, "Print version and exit")
	v       = flag.Bool("v", false, "Print version and exit")
)

func main() {
	branch := ""
	if gitBranch != "master" && gitBranch != "main" {
		branch = "- branch: " + gitBranch + " "
	}

	// Load the environment variables
	getEnv()

	flag.Parse()
	if *Version || *v {
		// Version contains version and Git commit information.
		//
		// The placeholders are replaced on `git archive` using the `export-subst` attribute.
		var intVersion = version.Version(appFriendlyName, fmt.Sprintf("%s (%s) %s", gitTag, gitRevision, branch), "$Format:%(describe)$", "$Format:%H$")

		intVersion.Print(maker)
		os.Exit(0)
	}

	log.Infof("%s by %s, version %s (%s) %s", appFriendlyName, maker, gitTag, gitRevision, branch)

	// Create a new http client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Playback client tolerates self-signed certificates (internal MediaMTX service)
	playbackClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Tipp: Potentially use the chi router here if you want to work nicer than the build in http tooling
	// Start the application
	var router http.ServeMux
	setupRoutes(&router, client, playbackClient)

	log.Info("Starting MediaMTX Path Viewer")
	log.Infof("Listening on port %s, via link: http://localhost:%s%s", port, port, basePath)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), &router); err != http.ErrServerClosed {
		// unexpected error. socket in use?
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func setupRoutes(router *http.ServeMux, client *http.Client, playbackClient *http.Client) {
	// Load the HTML templates
	indexHTML, err := res.ReadFile("static/html_templates/index.html")
	log.Should(err)

	pathsHTML, err := res.ReadFile("static/html_templates/path_list.html")
	log.Should(err)

	detailHTML, err := res.ReadFile("static/html_templates/stream_detail.html")
	log.Should(err)

	recordingsPageHTML, err := res.ReadFile("static/html_templates/recordings.html")
	log.Should(err)

	recordingsListHTML, err := res.ReadFile("static/html_templates/recordings_list.html")
	log.Should(err)

	recordingDetailHTML, err := res.ReadFile("static/html_templates/recording_detail.html")
	log.Should(err)

	monitoringPageHTML, err := res.ReadFile("static/html_templates/monitoring.html")
	log.Should(err)

	monitoringListHTML, err := res.ReadFile("static/html_templates/monitoring_list.html")
	log.Should(err)

	// Version string for display
	versionStr := gitTag
	if versionStr == "" {
		versionStr = gitRevision
	}

	router.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		htmlData := HTMLdata{
			BaseURL:   basePath,
			PageTitle: "MediaMTX Path Viewer",
			Version:   versionStr,
		}

		temp := template.Must(template.New("index").Parse(string(indexHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Update the view count for a path
	router.HandleFunc(basePath+"/viewCount/{id}", func(w http.ResponseWriter, r *http.Request) {
		// log.Infof("HTMX received: viewCount %s %s", r.PathValue("id"), r.Header.Get("HX-Request"))

		ID := r.PathValue("id")
		ID = strings.ReplaceAll(ID, "-", "/")

		// Get the paths from the MediaMTX server
		MediaMTX_Data, err := getMediamtxPath(client, ID) // Get the first 100 paths
		if err != nil {
			log.Errorf("Error getting paths from MediaMTX: %s", err)
			http.Error(w, "Error getting paths from MediaMTX", http.StatusInternalServerError)
			return
		}

		type viewers struct {
			BaseURL      string
			Error        string
			PageTitle    string
			TotalReaders int
		}
		viewCounter := viewers{
			BaseURL:      basePath,
			PageTitle:    "View Counter",
			TotalReaders: MediaMTX_Data.TotalReaders,
		}

		// Load the server paths template

		temp := template.Must(template.New("viewCounter/" + MediaMTX_Data.ID).Parse("Viewers: {{.TotalReaders}}"))
		err = temp.Execute(w, viewCounter)
		log.Should(err)
	})

	// Handle Connect to device request
	router.HandleFunc(basePath+"/connect-to-server/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTMX received: connect-to-server %s", r.Header.Get("HX-Request"))

		// Redirect if not an htmx request
		if r.Header.Get("HX-Request") != "true" {
			http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
			return
		}

		// Get the paths from the MediaMTX server
		MediaMTX_Data, err := getMediamtxPaths(client, 0, 100) // Get the first 100 paths
		if err != nil {
			log.Errorf("Error getting paths from MediaMTX: %s", err)
			http.Error(w, "Error getting paths from MediaMTX", http.StatusInternalServerError)
			return
		}

		// Cross-reference with recordings to mark which streams are recording
		recordings, recErr := getMediamtxRecordings(client, 0, 100)
		if recErr == nil {
			recSet := make(map[string]bool, len(recordings.Items))
			for _, rec := range recordings.Items {
				recSet[rec.Name] = true
			}
			for i, path := range MediaMTX_Data.Items {
				if recSet[path.Name] {
					MediaMTX_Data.Items[i].IsRecording = true
				}
			}
		}

		htmlData := HTMLdata{
			BaseURL:   basePath,
			PageTitle: "Server Paths",
			ItemCount: MediaMTX_Data.ItemCount,
			PageCount: MediaMTX_Data.PageCount,
			Items:     MediaMTX_Data.Items,
			Groups:    groupPaths(MediaMTX_Data.Items),
		}
		// Load the server paths template
		temp := template.Must(template.New("serverPaths").Parse(string(pathsHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Stream detail for modal view
	router.HandleFunc(basePath+"/stream-detail/{id}", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTMX received: stream-detail %s %s", r.PathValue("id"), r.Header.Get("HX-Request"))

		ID := r.PathValue("id")
		ID = strings.ReplaceAll(ID, "-", "/")

		MediaMTX_Data, err := getMediamtxPath(client, ID)
		if err != nil {
			log.Errorf("Error getting path from MediaMTX: %s", err)
			http.Error(w, "Error getting path from MediaMTX", http.StatusInternalServerError)
			return
		}

		// Check if this stream has recordings
		recordings, recErr := getMediamtxRecordings(client, 0, 100)
		if recErr == nil {
			for _, rec := range recordings.Items {
				if rec.Name == MediaMTX_Data.Name {
					MediaMTX_Data.IsRecording = true
					break
				}
			}
		}

		// Add BaseURL for template use
		type detailData struct {
			Path
			BaseURL string
		}

		data := detailData{
			Path:    MediaMTX_Data,
			BaseURL: basePath,
		}

		temp := template.Must(template.New("streamDetail").Parse(string(detailHTML)))
		err = temp.Execute(w, data)
		log.Should(err)
	})

	// Recordings page
	router.HandleFunc(basePath+"/recordings/", func(w http.ResponseWriter, r *http.Request) {
		htmlData := RecordingsHTMLdata{
			BaseURL:   basePath,
			PageTitle: "Recordings - MediaMTX Path Viewer",
			Version:   versionStr,
		}

		temp := template.Must(template.New("recordings").Parse(string(recordingsPageHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Recordings list HTMX endpoint
	router.HandleFunc(basePath+"/recordings-list/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTMX received: recordings-list %s", r.Header.Get("HX-Request"))

		if r.Header.Get("HX-Request") != "true" {
			http.Redirect(w, r, basePath+"/recordings/", http.StatusSeeOther)
			return
		}

		recordings, err := getMediamtxRecordings(client, 0, 100)
		if err != nil {
			log.Errorf("Error getting recordings from MediaMTX: %s", err)
			http.Error(w, "Error getting recordings from MediaMTX", http.StatusInternalServerError)
			return
		}

		htmlData := RecordingsHTMLdata{
			BaseURL:     basePath,
			PageTitle:   "Recordings",
			ItemCount:   recordings.ItemCount,
			PageCount:   recordings.PageCount,
			Items:       recordings.Items,
			Groups:      groupRecordings(recordings.Items),
			PlaybackURL: MEDIAMTX_PLAYBACK_URL,
		}

		temp := template.Must(template.New("recordingsList").Parse(string(recordingsListHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Recording detail for modal view
	router.HandleFunc(basePath+"/recording-detail/{id}", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTMX received: recording-detail %s %s", r.PathValue("id"), r.Header.Get("HX-Request"))

		ID := r.PathValue("id")
		name := strings.ReplaceAll(ID, "-", "/")

		// Find the recording from the list
		recordings, err := getMediamtxRecordings(client, 0, 100)
		if err != nil {
			log.Errorf("Error getting recordings from MediaMTX: %s", err)
			http.Error(w, "Error getting recordings from MediaMTX", http.StatusInternalServerError)
			return
		}

		var recording Recording
		found := false
		for _, rec := range recordings.Items {
			if rec.Name == name {
				recording = rec
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "Recording not found", http.StatusNotFound)
			return
		}

		// Fetch playback segments if playback URL is configured
		var segments []PlaybackSegment
		if MEDIAMTX_PLAYBACK_URL != "" {
			segments, err = getPlaybackSegments(playbackClient, name)
			if err != nil {
				log.Errorf("Error getting playback segments: %s", err)
				// Continue without segments — still show metadata
			}
		}

		data := RecordingDetailData{
			Recording:   recording,
			BaseURL:     basePath,
			PlaybackURL: MEDIAMTX_PLAYBACK_URL,
			Segments:    segments,
		}

		temp := template.Must(template.New("recordingDetail").Parse(string(recordingDetailHTML)))
		err = temp.Execute(w, data)
		log.Should(err)
	})

	// Monitoring page
	router.HandleFunc(basePath+"/monitoring/", func(w http.ResponseWriter, r *http.Request) {
		htmlData := MonitoringHTMLdata{
			BaseURL:   basePath,
			PageTitle: "Monitoring - MediaMTX Path Viewer",
			Version:   versionStr,
		}

		temp := template.Must(template.New("monitoring").Parse(string(monitoringPageHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Monitoring list HTMX endpoint
	router.HandleFunc(basePath+"/monitoring-list/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTMX received: monitoring-list %s", r.Header.Get("HX-Request"))

		if r.Header.Get("HX-Request") != "true" {
			http.Redirect(w, r, basePath+"/monitoring/", http.StatusSeeOther)
			return
		}

		// Fetch all monitoring data (errors are non-fatal; partial data is OK)
		serverInfo, infoErr := getMediamtxInfo(client)
		if infoErr != nil {
			log.Errorf("Error getting server info: %s", infoErr)
		}

		paths, pathErr := getMediamtxPaths(client, 0, 100)
		if pathErr != nil {
			log.Errorf("Error getting paths: %s", pathErr)
		}

		webrtcSessions, webrtcErr := getMediamtxWebRTCSessions(client, 0, 100)
		if webrtcErr != nil {
			log.Errorf("Error getting WebRTC sessions: %s", webrtcErr)
		}

		rtspSessions, rtspErr := getMediamtxRTSPSessions(client, 0, 100)
		if rtspErr != nil {
			log.Errorf("Error getting RTSP sessions: %s", rtspErr)
		}

		// Merge TLS (RTSPS) sessions into the RTSP list so counts/geo/summaries include them.
		// A 404 here just means the RTSPS service is disabled on the server, so log at debug.
		if rtspsSessions, e := getMediamtxRTSPSSessions(client, 0, 100); e != nil {
			log.Debugf("Skipping RTSPS sessions: %s", e)
		} else {
			rtspSessions.Items = append(rtspSessions.Items, rtspsSessions.Items...)
			rtspSessions.ItemCount += rtspsSessions.ItemCount
		}

		rtmpConns, rtmpErr := getMediamtxRTMPConns(client, 0, 100)
		if rtmpErr != nil {
			log.Errorf("Error getting RTMP connections: %s", rtmpErr)
		}

		// Merge TLS (RTMPS) connections into the RTMP list (debug-log if the service is disabled).
		if rtmpsConns, e := getMediamtxRTMPSConns(client, 0, 100); e != nil {
			log.Debugf("Skipping RTMPS connections: %s", e)
		} else {
			rtmpConns.Items = append(rtmpConns.Items, rtmpsConns.Items...)
			rtmpConns.ItemCount += rtmpsConns.ItemCount
		}

		hlsMuxers, hlsErr := getMediamtxHLSMuxers(client, 0, 100)
		if hlsErr != nil {
			log.Errorf("Error getting HLS muxers: %s", hlsErr)
		}

		hlsSessions, hlsSessErr := getMediamtxHLSSessions(client, 0, 100)
		if hlsSessErr != nil {
			log.Errorf("Error getting HLS sessions: %s", hlsSessErr)
		}

		srtConns, srtErr := getMediamtxSRTConns(client, 0, 100)
		if srtErr != nil {
			log.Errorf("Error getting SRT connections: %s", srtErr)
		}

		// MoQ (Media-over-QUIC) may be disabled on the server; a 404 is expected there.
		moqSessions, moqErr := getMediamtxMoQSessions(client, 0, 100)
		if moqErr != nil {
			log.Debugf("Skipping MoQ sessions: %s", moqErr)
		}

		// Build per-stream summaries
		summaries := buildStreamSummaries(
			paths.Items,
			webrtcSessions.Items,
			rtspSessions.Items,
			rtmpConns.Items,
			hlsMuxers.Items,
			srtConns.Items,
			hlsSessions.Items,
			moqSessions.Items,
		)

		// GeoIP lookup for viewer countries
		remoteIPs := collectRemoteIPs(webrtcSessions.Items, rtspSessions.Items, rtmpConns.Items, srtConns.Items, hlsSessions.Items, moqSessions.Items)
		geoMap := lookupGeoIP(client, remoteIPs)
		countrySummaries := applyGeoData(geoMap, webrtcSessions.Items, rtspSessions.Items, rtmpConns.Items, srtConns.Items, hlsSessions.Items, moqSessions.Items)

		// Calculate totals
		var totalViewers int
		var totalBytes uint64
		for _, s := range summaries {
			totalViewers += s.TotalViewers
			totalBytes += s.TotalBytes
		}

		htmlData := MonitoringHTMLdata{
			BaseURL:           basePath,
			PageTitle:         "Monitoring",
			ServerInfo:        serverInfo,
			TotalStreams:      paths.ItemCount,
			TotalViewers:      totalViewers,
			TotalBandwidthStr: formatBytes(totalBytes),
			WebRTCCount:       webrtcSessions.ItemCount,
			RTSPCount:         rtspSessions.ItemCount,
			RTMPCount:         rtmpConns.ItemCount,
			HLSCount:          hlsMuxers.ItemCount,
			HLSSessionCount:   hlsSessions.ItemCount,
			SRTCount:          srtConns.ItemCount,
			MoQCount:          moqSessions.ItemCount,
			StreamSummaries:   summaries,
			CountrySummaries:  countrySummaries,
			WebRTCSessions:    webrtcSessions.Items,
			RTSPSessions:      rtspSessions.Items,
			RTMPConns:         rtmpConns.Items,
			HLSMuxers:         hlsMuxers.Items,
			HLSSessions:       hlsSessions.Items,
			SRTConns:          srtConns.Items,
			MoQSessions:       moqSessions.Items,
		}

		temp := template.Must(template.New("monitoringList").Parse(string(monitoringListHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Serve the static CSS and JS files
	serveStatic(router)
}

func serveStatic(router *http.ServeMux) {
	// Server Icon
	router.HandleFunc(basePath+"/static/pictures/icon.png", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/pictures/icon.png")
		log.Should(err)
		http.ServeContent(w, r, "icon.png", time.Now(), bytes.NewReader(file))
	})

	// Serve the static CSS and JS files
	router.HandleFunc(basePath+"/static/css/bootstrap_5_3_3.min.css", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/css/bootstrap_5_3_3.min.css")
		log.Should(err)
		http.ServeContent(w, r, "bootstrap_5_3_3.min.css", time.Now(), bytes.NewReader(file))
	})
	router.HandleFunc(basePath+"/static/css/custom.css", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/css/custom.css")
		log.Should(err)
		http.ServeContent(w, r, "custom.css", time.Now(), bytes.NewReader(file))
	})
	router.HandleFunc(basePath+"/static/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/js/htmx.min.js")
		log.Should(err)
		http.ServeContent(w, r, "htmx.min.js", time.Now(), bytes.NewReader(file))
	})
	router.HandleFunc(basePath+"/static/js/hls.js", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/js/hls.js")
		log.Should(err)
		http.ServeContent(w, r, "hls.js", time.Now(), bytes.NewReader(file))
	})
	router.HandleFunc(basePath+"/static/js/bootstrap.bundle.min.js", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/js/bootstrap.bundle.min.js")
		log.Should(err)
		http.ServeContent(w, r, "bootstrap.bundle.min.js", time.Now(), bytes.NewReader(file))
	})
}

func getEnv() {
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")
	if err != nil {
		log.Warnf("Error loading .env file, %s, ignore this if ran in docker", err)
	}

	MEDIAMTX_API_URL = os.Getenv("MEDIAMTX_API_URL")
	MEDIAMTX_API_PORT = os.Getenv("MEDIAMTX_API_PORT")
	MEDIAMTX_USERNAME = os.Getenv("MEDIAMTX_USERNAME")
	MEDIAMTX_PASSWORD = os.Getenv("MEDIAMTX_PASSWORD")

	MEDIAMTX_WEBRTC_URL = os.Getenv("MEDIAMTX_WEBRTC_URL")
	MEDIAMTX_HLS_URL = os.Getenv("MEDIAMTX_HLS_URL")
	MEDIAMTX_RTMP_URL = os.Getenv("MEDIAMTX_RTMP_URL")
	MEDIAMTX_RTSP_URL = os.Getenv("MEDIAMTX_RTSP_URL")
	MEDIAMTX_MOQ_URL = os.Getenv("MEDIAMTX_MOQ_URL")
	MEDIAMTX_PLAYBACK_URL = os.Getenv("MEDIAMTX_PLAYBACK_URL")

	APP_PORT := os.Getenv("APP_PORT")
	if APP_PORT != "" {
		port = APP_PORT
	}

	APP_PATH := os.Getenv("APP_PATH")
	if APP_PATH != "" {
		basePath = APP_PATH
	}

	// Check if the basePath is set correctly
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if basePath != "" && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimSuffix(basePath, "/")
	}

	// Check if the MediaMTX Host is defined
	if MEDIAMTX_API_URL == "" || MEDIAMTX_WEBRTC_URL == "" || MEDIAMTX_HLS_URL == "" {
		log.Fatalf("Missing required environment variables: MEDIAMTX_API_URL, MEDIAMTX_WEBRTC_URL, and MEDIAMTX_HLS_URL must all be set")
	}
	if MEDIAMTX_USERNAME == "" || MEDIAMTX_PASSWORD == "" {
		log.Infof("No MEDIAMTX_USERNAME or MEDIAMTX_PASSWORD defined, no authentication will be used")
	}
	if MEDIAMTX_PLAYBACK_URL != "" {
		log.Infof("Playback server configured at %s", MEDIAMTX_PLAYBACK_URL)
	}

	// Force set the MediaMTX Host to http
	MEDIAMTX_API_URL = strings.Replace(MEDIAMTX_API_URL, "https://", "http://", 1)
	if !strings.HasPrefix(MEDIAMTX_API_URL, "http://") {
		MEDIAMTX_API_URL = "http://" + MEDIAMTX_API_URL
	}

	// Set the default port if not defined
	if MEDIAMTX_API_PORT == "" {
		MEDIAMTX_API_PORT = "9997"
	}

}
