package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
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
