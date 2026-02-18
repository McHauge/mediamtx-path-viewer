package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

func getMediamtxPaths(client *http.Client, page, itemsPrePage int) (MediaMTX, error) {
	data := MediaMTX{}

	host := MEDIAMTX_API_URL + ":" + MEDIAMTX_API_PORT
	url := fmt.Sprintf("%s/v3/paths/list/?page=%d&itemsPerPage=%d", host, page, itemsPrePage)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data, err
	}

	if MEDIAMTX_USERNAME != "" || MEDIAMTX_PASSWORD != "" {
		req.SetBasicAuth(MEDIAMTX_USERNAME, MEDIAMTX_PASSWORD)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return data, fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	// Decode the JSON response
	err = json.Unmarshal([]byte(respBody), &data)
	if err != nil {
		return data, err
	}

	// Add the ReadyTimeStr
	for i, path := range data.Items {
		path.ReadyTimeStr = ""

		// If the ReadyTime is not set, skip it
		if path.ReadyTime == (time.Time{}) {
			continue
		}
		path.ReadyTimeStr = path.ReadyTime.Format("2006-01-02 15:04:05")
		data.Items[i] = path
	}

	for i, path := range data.Items {
		path.ID = strings.ReplaceAll(path.ConfName, "/", "-")

		// Format the path data
		path = formatPathData(path)

		// Add the total readers
		path.TotalReaders = len(path.Readers)
		data.Items[i] = path
	}

	data.Items = sortPaths(data.Items)

	return data, err
}

func getMediamtxPath(client *http.Client, path string) (Path, error) {
	data := Path{}

	host := MEDIAMTX_API_URL + ":" + MEDIAMTX_API_PORT
	url := fmt.Sprintf("%s/v3/paths/get/%s", host, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data, err
	}

	if MEDIAMTX_USERNAME != "" || MEDIAMTX_PASSWORD != "" {
		req.SetBasicAuth(MEDIAMTX_USERNAME, MEDIAMTX_PASSWORD)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return data, fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	// Decode the JSON response
	err = json.Unmarshal([]byte(respBody), &data)
	if err != nil {
		return data, err
	}

	// Add the ReadyTimeStr
	data.ReadyTimeStr = ""

	// If the ReadyTime is not set, skip it
	if data.ReadyTime != (time.Time{}) {
		data.ReadyTimeStr = data.ReadyTime.Format("2006-01-02 15:04:05")
	}

	// Format the path data
	data = formatPathData(data)

	// Add the total readers
	data.TotalReaders = len(data.Readers)

	return data, err
}

func formatPathData(path Path) Path {
	path.ID = strings.ReplaceAll(path.Name, "/", "-")

	if MEDIAMTX_WEBRTC_URL != "" {
		path.StreamWebRTC = fmt.Sprintf("%s/%s", MEDIAMTX_WEBRTC_URL, path.Name)
	}
	if MEDIAMTX_HLS_URL != "" {
		path.StreamHls = fmt.Sprintf("%s/%s", MEDIAMTX_HLS_URL, path.Name)
	}
	if MEDIAMTX_RTMP_URL != "" {
		path.StreamRtmp = fmt.Sprintf("%s/%s", MEDIAMTX_RTMP_URL, path.Name)
	}
	if MEDIAMTX_RTSP_URL != "" {
		path.StreamRtsp = fmt.Sprintf("%s/%s", MEDIAMTX_RTSP_URL, path.Name)
	}

	// Add the stream URL
	if path.Source.Type == "webRTCSession" {
		path.StreamUrl = fmt.Sprintf("%s/%s", MEDIAMTX_WEBRTC_URL, path.Name)
	} else {
		path.StreamUrl = fmt.Sprintf("%s/%s", MEDIAMTX_HLS_URL, path.Name)
	}

	// Pretty Name & Path
	x := strings.Split(path.Name, "/")
	for i := 0; i < len(x); i++ {
		if len(x[i]) == 0 {
			continue
		}
		x[i] = strings.ToUpper(string(x[i][0])) + x[i][1:] // only first letter upper case
	}
	if len(x) > 1 {
		path.PrettyName = x[len(x)-1]
		path.PathName = strings.Join(x[:len(x)-1], "/")
	} else {
		path.PrettyName = x[0]
		path.PathName = "" // no path / using root path
	}

	if strings.Contains(path.PrettyName, "_") {
		x := strings.Split(path.PrettyName, "_")
		for i := 0; i < len(x); i++ {
			if len(x[i]) == 0 {
				continue
			}
			x[i] = strings.ToUpper(string(x[i][0])) + x[i][1:] // only first letter upper case
		}
		path.PrettyName = strings.Join(x, " ")
	}

	return path
}

// sortPaths takes in a list of paths and then first sorts them by the root path, then sub path and then by the name
func sortPaths(paths []Path) []Path {
	// Sort the paths by the PathName and then by the Name
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].PathName == paths[j].PathName {
			return paths[i].PrettyName < paths[j].PrettyName
		}
		return paths[i].PathName < paths[j].PathName
	})
	return paths
}

// Recording API functions

func getMediamtxRecordings(client *http.Client, page, itemsPerPage int) (RecordingList, error) {
	data := RecordingList{}

	host := MEDIAMTX_API_URL + ":" + MEDIAMTX_API_PORT
	url := fmt.Sprintf("%s/v3/recordings/list?page=%d&itemsPerPage=%d", host, page, itemsPerPage)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data, err
	}

	if MEDIAMTX_USERNAME != "" || MEDIAMTX_PASSWORD != "" {
		req.SetBasicAuth(MEDIAMTX_USERNAME, MEDIAMTX_PASSWORD)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return data, fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	// log.Infof("Raw recordings response: %s", string(respBody))

	err = json.Unmarshal(respBody, &data)
	if err != nil {
		return data, err
	}

	for i, rec := range data.Items {
		data.Items[i] = formatRecordingData(rec)
	}

	data.Items = sortRecordings(data.Items)

	return data, nil
}

func getPlaybackSegments(client *http.Client, pathName string) ([]PlaybackSegment, error) {
	if MEDIAMTX_PLAYBACK_URL == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/list?path=%s", MEDIAMTX_PLAYBACK_URL, pathName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Playback server returned status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// log.Infof("Raw recordings response: %s", string(respBody))

	var segments []PlaybackSegment
	err = json.Unmarshal(respBody, &segments)
	if err != nil {
		return nil, err
	}

	for i, seg := range segments {
		segments[i].StartStr = seg.Start.Format("2006-01-02 15:04:05")
		segments[i].StartRFC = seg.Start.UTC().Format(time.RFC3339)

		totalSeconds := int(seg.Duration)
		hours := totalSeconds / 3600
		minutes := (totalSeconds % 3600) / 60
		seconds := totalSeconds % 60
		if hours > 0 {
			segments[i].DurationStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
		} else if minutes > 0 {
			segments[i].DurationStr = fmt.Sprintf("%dm %ds", minutes, seconds)
		} else {
			segments[i].DurationStr = fmt.Sprintf("%ds", seconds)
		}
	}

	return segments, nil
}

func formatRecordingData(rec Recording) Recording {
	rec.ID = strings.ReplaceAll(rec.Name, "/", "-")
	rec.SegmentCount = len(rec.Segments)

	// Pretty Name & Path (same logic as formatPathData)
	x := strings.Split(rec.Name, "/")
	for i := 0; i < len(x); i++ {
		if len(x[i]) == 0 {
			continue
		}
		x[i] = strings.ToUpper(string(x[i][0])) + x[i][1:]
	}
	if len(x) > 1 {
		rec.PrettyName = x[len(x)-1]
		rec.PathName = strings.Join(x[:len(x)-1], "/")
	} else {
		rec.PrettyName = x[0]
		rec.PathName = ""
	}

	if strings.Contains(rec.PrettyName, "_") {
		parts := strings.Split(rec.PrettyName, "_")
		for i := 0; i < len(parts); i++ {
			if len(parts[i]) == 0 {
				continue
			}
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
		rec.PrettyName = strings.Join(parts, " ")
	}

	// Segment timestamps
	if len(rec.Segments) > 0 {
		rec.OldestSegment = rec.Segments[0].Start.Format("2006-01-02 15:04:05")
		rec.NewestSegment = rec.Segments[len(rec.Segments)-1].Start.Format("2006-01-02 15:04:05")
	}

	return rec
}

func sortRecordings(recordings []Recording) []Recording {
	sort.Slice(recordings, func(i, j int) bool {
		if recordings[i].PathName == recordings[j].PathName {
			return recordings[i].PrettyName < recordings[j].PrettyName
		}
		return recordings[i].PathName < recordings[j].PathName
	})
	return recordings
}

func groupRecordings(recordings []Recording) []RecordingGroup {
	groupMap := make(map[string][]Recording)
	var groupOrder []string

	for _, r := range recordings {
		name := r.PathName
		if name == "" {
			name = "Recordings"
		}
		if _, exists := groupMap[name]; !exists {
			groupOrder = append(groupOrder, name)
		}
		groupMap[name] = append(groupMap[name], r)
	}

	groups := make([]RecordingGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		groups = append(groups, RecordingGroup{
			GroupName:  name,
			Recordings: groupMap[name],
		})
	}
	return groups
}

// Monitoring API functions

// GeoIP cache
var (
	geoCache   = make(map[string]GeoIPResult)
	geoCacheMu sync.RWMutex
)

// extractIP strips the port from a remoteAddr string (e.g., "1.2.3.4:5678" -> "1.2.3.4")
func extractIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// lookupGeoIP does a batch GeoIP lookup via ip-api.com for any IPs not already cached.
// It returns a map from IP to GeoIPResult.
func lookupGeoIP(client *http.Client, ips []string) map[string]GeoIPResult {
	result := make(map[string]GeoIPResult)

	// Collect unique IPs that need lookup
	var toLookup []string
	seen := make(map[string]bool)

	geoCacheMu.RLock()
	for _, ip := range ips {
		if ip == "" || seen[ip] {
			continue
		}
		seen[ip] = true
		if cached, ok := geoCache[ip]; ok {
			result[ip] = cached
		} else {
			toLookup = append(toLookup, ip)
		}
	}
	geoCacheMu.RUnlock()

	if len(toLookup) == 0 {
		return result
	}

	// Build batch request (ip-api.com allows up to 100 per batch)
	type batchQuery struct {
		Query  string `json:"query"`
		Fields string `json:"fields"`
	}
	batch := make([]batchQuery, 0, len(toLookup))
	for _, ip := range toLookup {
		batch = append(batch, batchQuery{
			Query:  ip,
			Fields: "status,country,countryCode,query",
		})
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return result
	}

	req, err := http.NewRequest("POST", "http://ip-api.com/batch?fields=status,country,countryCode,query", bytes.NewReader(body))
	if err != nil {
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result
	}

	var results []GeoIPResult
	if err := json.Unmarshal(respBody, &results); err != nil {
		return result
	}

	// Update cache and result map
	geoCacheMu.Lock()
	for _, r := range results {
		if r.Status == "success" {
			geoCache[r.Query] = r
			result[r.Query] = r
		}
	}
	geoCacheMu.Unlock()

	return result
}

// collectRemoteIPs extracts unique IPs from all connection types
func collectRemoteIPs(
	webrtcSessions []WebRTCSession,
	rtspSessions []RTSPSession,
	rtmpConns []RTMPConn,
	srtConns []SRTConn,
) []string {
	seen := make(map[string]bool)
	var ips []string

	addIP := func(remoteAddr string) {
		ip := extractIP(remoteAddr)
		if ip != "" && !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}

	for _, s := range webrtcSessions {
		addIP(s.RemoteAddr)
	}
	for _, s := range rtspSessions {
		addIP(s.RemoteAddr)
	}
	for _, c := range rtmpConns {
		addIP(c.RemoteAddr)
	}
	for _, c := range srtConns {
		addIP(c.RemoteAddr)
	}

	return ips
}

// applyGeoData enriches connection structs with country info and returns a country summary
func applyGeoData(
	geoMap map[string]GeoIPResult,
	webrtcSessions []WebRTCSession,
	rtspSessions []RTSPSession,
	rtmpConns []RTMPConn,
	srtConns []SRTConn,
) []CountrySummary {
	countryCount := make(map[string]*CountrySummary)
	var countryOrder []string

	addCountry := func(ip string) {
		geo, ok := geoMap[ip]
		if !ok || geo.Country == "" {
			return
		}
		if _, exists := countryCount[geo.CountryCode]; !exists {
			countryCount[geo.CountryCode] = &CountrySummary{
				Country:     geo.Country,
				CountryCode: geo.CountryCode,
			}
			countryOrder = append(countryOrder, geo.CountryCode)
		}
		countryCount[geo.CountryCode].Count++
	}

	for i, s := range webrtcSessions {
		ip := extractIP(s.RemoteAddr)
		if geo, ok := geoMap[ip]; ok {
			webrtcSessions[i].Country = geo.Country
			webrtcSessions[i].CountryCode = geo.CountryCode
		}
		if s.State == "read" {
			addCountry(ip)
		}
	}
	for i, s := range rtspSessions {
		ip := extractIP(s.RemoteAddr)
		if geo, ok := geoMap[ip]; ok {
			rtspSessions[i].Country = geo.Country
			rtspSessions[i].CountryCode = geo.CountryCode
		}
		if s.State == "read" {
			addCountry(ip)
		}
	}
	for i, c := range rtmpConns {
		ip := extractIP(c.RemoteAddr)
		if geo, ok := geoMap[ip]; ok {
			rtmpConns[i].Country = geo.Country
			rtmpConns[i].CountryCode = geo.CountryCode
		}
		if c.State == "read" {
			addCountry(ip)
		}
	}
	for i, c := range srtConns {
		ip := extractIP(c.RemoteAddr)
		if geo, ok := geoMap[ip]; ok {
			srtConns[i].Country = geo.Country
			srtConns[i].CountryCode = geo.CountryCode
		}
		if c.State == "read" {
			addCountry(ip)
		}
	}

	// Build sorted summary (by count descending)
	summaries := make([]CountrySummary, 0, len(countryOrder))
	for _, code := range countryOrder {
		summaries = append(summaries, *countryCount[code])
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Count > summaries[j].Count
	})

	return summaries
}

func formatBytes(b uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func mediamtxAPIRequest(client *http.Client, endpoint string, target interface{}) error {
	host := MEDIAMTX_API_URL + ":" + MEDIAMTX_API_PORT
	url := host + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if MEDIAMTX_USERNAME != "" || MEDIAMTX_PASSWORD != "" {
		req.SetBasicAuth(MEDIAMTX_USERNAME, MEDIAMTX_PASSWORD)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("MediaMTX API returned status %d: %s", resp.StatusCode, string(body))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, target)
}

func getMediamtxInfo(client *http.Client) (ServerInfo, error) {
	data := ServerInfo{}
	err := mediamtxAPIRequest(client, "/v3/info", &data)
	if err != nil {
		return data, err
	}

	started, parseErr := time.Parse(time.RFC3339Nano, data.Started)
	if parseErr == nil {
		uptime := time.Since(started)
		days := int(uptime.Hours()) / 24
		hours := int(uptime.Hours()) % 24
		minutes := int(uptime.Minutes()) % 60
		if days > 0 {
			data.UptimeStr = fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
		} else if hours > 0 {
			data.UptimeStr = fmt.Sprintf("%dh %dm", hours, minutes)
		} else {
			data.UptimeStr = fmt.Sprintf("%dm", minutes)
		}
	} else {
		data.UptimeStr = "Unknown"
	}

	return data, nil
}

func getMediamtxHLSMuxers(client *http.Client, page, itemsPerPage int) (HLSMuxerList, error) {
	data := HLSMuxerList{}
	endpoint := fmt.Sprintf("/v3/hlsmuxers/list?page=%d&itemsPerPage=%d", page, itemsPerPage)
	err := mediamtxAPIRequest(client, endpoint, &data)
	if err != nil {
		return data, err
	}

	for i, muxer := range data.Items {
		data.Items[i].BytesSentStr = formatBytes(muxer.BytesSent)
		if t, e := time.Parse(time.RFC3339Nano, muxer.Created); e == nil {
			data.Items[i].CreatedStr = t.Format("2006-01-02 15:04:05")
		}
		if t, e := time.Parse(time.RFC3339Nano, muxer.LastRequest); e == nil {
			data.Items[i].LastRequestStr = t.Format("2006-01-02 15:04:05")
		}
	}

	return data, nil
}

func getMediamtxRTSPSessions(client *http.Client, page, itemsPerPage int) (RTSPSessionList, error) {
	data := RTSPSessionList{}
	endpoint := fmt.Sprintf("/v3/rtspsessions/list?page=%d&itemsPerPage=%d", page, itemsPerPage)
	err := mediamtxAPIRequest(client, endpoint, &data)
	if err != nil {
		return data, err
	}

	for i, sess := range data.Items {
		data.Items[i].BytesReceivedStr = formatBytes(sess.BytesReceived)
		data.Items[i].BytesSentStr = formatBytes(sess.BytesSent)
		data.Items[i].RTPPacketsJitterStr = fmt.Sprintf("%.2f ms", sess.RTPPacketsJitter)
	}

	return data, nil
}

func getMediamtxRTMPConns(client *http.Client, page, itemsPerPage int) (RTMPConnList, error) {
	data := RTMPConnList{}
	endpoint := fmt.Sprintf("/v3/rtmpconns/list?page=%d&itemsPerPage=%d", page, itemsPerPage)
	err := mediamtxAPIRequest(client, endpoint, &data)
	if err != nil {
		return data, err
	}

	for i, conn := range data.Items {
		data.Items[i].BytesReceivedStr = formatBytes(conn.BytesReceived)
		data.Items[i].BytesSentStr = formatBytes(conn.BytesSent)
	}

	return data, nil
}

func getMediamtxSRTConns(client *http.Client, page, itemsPerPage int) (SRTConnList, error) {
	data := SRTConnList{}
	endpoint := fmt.Sprintf("/v3/srtconns/list?page=%d&itemsPerPage=%d", page, itemsPerPage)
	err := mediamtxAPIRequest(client, endpoint, &data)
	if err != nil {
		return data, err
	}

	for i, conn := range data.Items {
		data.Items[i].BytesReceivedStr = formatBytes(conn.BytesReceived)
		data.Items[i].BytesSentStr = formatBytes(conn.BytesSent)
	}

	return data, nil
}

func getMediamtxWebRTCSessions(client *http.Client, page, itemsPerPage int) (WebRTCSessionList, error) {
	data := WebRTCSessionList{}
	endpoint := fmt.Sprintf("/v3/webrtcsessions/list?page=%d&itemsPerPage=%d", page, itemsPerPage)
	err := mediamtxAPIRequest(client, endpoint, &data)
	if err != nil {
		return data, err
	}

	for i, sess := range data.Items {
		data.Items[i].BytesReceivedStr = formatBytes(sess.BytesReceived)
		data.Items[i].BytesSentStr = formatBytes(sess.BytesSent)
		data.Items[i].RTPPacketsJitterStr = fmt.Sprintf("%.2f ms", sess.RTPPacketsJitter)
	}

	return data, nil
}

func buildStreamSummaries(
	paths []Path,
	webrtcSessions []WebRTCSession,
	rtspSessions []RTSPSession,
	rtmpConns []RTMPConn,
	hlsMuxers []HLSMuxer,
	srtConns []SRTConn,
) []StreamSummary {
	summaryMap := make(map[string]*StreamSummary)
	var order []string

	for _, p := range paths {
		s := &StreamSummary{
			Name:       p.Name,
			SourceType: p.Source.Type,
			Tracks:     p.Tracks,
		}
		summaryMap[p.Name] = s
		order = append(order, p.Name)
	}

	for _, sess := range webrtcSessions {
		if s, ok := summaryMap[sess.Path]; ok {
			if sess.State == "read" {
				s.WebRTCViewers++
			}
			s.TotalBytes += sess.BytesReceived + sess.BytesSent
		}
	}
	for _, sess := range rtspSessions {
		if s, ok := summaryMap[sess.Path]; ok {
			if sess.State == "read" {
				s.RTSPViewers++
			}
			s.TotalBytes += sess.BytesReceived + sess.BytesSent
		}
	}
	for _, conn := range rtmpConns {
		if s, ok := summaryMap[conn.Path]; ok {
			if conn.State == "read" {
				s.RTMPViewers++
			}
			s.TotalBytes += conn.BytesReceived + conn.BytesSent
		}
	}
	for _, muxer := range hlsMuxers {
		if s, ok := summaryMap[muxer.Path]; ok {
			s.HLSViewers++
			s.TotalBytes += muxer.BytesSent
		}
	}
	for _, conn := range srtConns {
		if s, ok := summaryMap[conn.Path]; ok {
			if conn.State == "read" {
				s.SRTViewers++
			}
			s.TotalBytes += conn.BytesReceived + conn.BytesSent
		}
	}

	result := make([]StreamSummary, 0, len(order))
	for _, name := range order {
		s := summaryMap[name]
		s.TotalViewers = s.WebRTCViewers + s.RTSPViewers + s.RTMPViewers + s.HLSViewers + s.SRTViewers
		s.BandwidthStr = formatBytes(s.TotalBytes)
		result = append(result, *s)
	}
	return result
}

// groupPaths takes a sorted list of paths and groups them by their PathName
func groupPaths(paths []Path) []PathGroup {
	groupMap := make(map[string][]Path)
	var groupOrder []string

	for _, p := range paths {
		name := p.PathName
		if name == "" {
			name = "Streams"
		}
		if _, exists := groupMap[name]; !exists {
			groupOrder = append(groupOrder, name)
		}
		groupMap[name] = append(groupMap[name], p)
	}

	groups := make([]PathGroup, 0, len(groupOrder))
	for _, name := range groupOrder {
		groups = append(groups, PathGroup{
			GroupName: name,
			Paths:     groupMap[name],
		})
	}
	return groups
}
