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
	// log.Warnf("Getting paths from %s", url)

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
		return data, err
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
	// log.Warnf("Getting paths from %s", url)

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
		return data, err
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
