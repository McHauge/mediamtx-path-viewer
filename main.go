package main

import (
	"bytes"
	"embed"
	_ "embed"
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
	MEDIAMTX_WEBRTC_URL string
	MEDIAMTX_HLS_URL    string
	MEDIAMTX_API_URL    string
	MEDIAMTX_API_PORT   string
	MEDIAMTX_USERNAME   string
	MEDIAMTX_PASSWORD   string

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

	// Tipp: Potentially use the chi router here if you want to work nicer than the build in http tooling
	// Start the application
	var router http.ServeMux
	setupRoutes(&router, client)

	log.Info("Starting MediaMTX Path Viewer")
	log.Infof("Listening on port %s, via link: http://localhost:%s%s", port, port, basePath)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), &router); err != http.ErrServerClosed {
		// unexpected error. socket in use?
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func setupRoutes(router *http.ServeMux, client *http.Client) {
	// Load the index.html template
	indexHTML, err := res.ReadFile("static/html_templates/index.html")
	log.Should(err)

	pathsHTML, err := res.ReadFile("static/html_templates/path_list.html")
	log.Should(err)

	router.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		htmlData := HTMLdata{
			BaseURL:   basePath,
			PageTitle: "MediaMTX Viewer",
		}

		temp := template.Must(template.New("index").Parse(string(indexHTML)))
		err = temp.Execute(w, htmlData)
		log.Should(err)
	})

	// Handle Connect to device request
	router.HandleFunc(basePath+"/connect-to-server/", func(w http.ResponseWriter, r *http.Request) {
		log.Warn("HTMX recived: connect-to-server ", r.Header.Get("HX-Request"))

		// Redirect if not an htmx request
		if r.Header.Get("HX-Request") != "true" {
			http.Redirect(w, r, basePath+"/", http.StatusSeeOther)
			return
		}

		// Get the paths from the MediaMTX server
		MediaMTX_Data, err := getMediamtxData(client, 0, 100) // Get the first 100 paths
		if err != nil {
			log.Errorf("Error getting paths from MediaMTX: %s", err)
			http.Error(w, "Error getting paths from MediaMTX", http.StatusInternalServerError)
			return
		}

		htmlData := HTMLdata{
			BaseURL:    basePath,
			PageTitle:  "Server Paths",
			ItemCount:  MediaMTX_Data.ItemCount,
			PageCount:  MediaMTX_Data.PageCount,
			Items:      MediaMTX_Data.Items,
		}
		// log.Warn(log.Indent(MediaMTX_Data))

		// Load the server paths template
		temp := template.Must(template.New("serverPaths").Parse(string(pathsHTML)))
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
	router.HandleFunc(basePath+"/static/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
		file, err := res.ReadFile("static/js/htmx.min.js")
		log.Should(err)
		http.ServeContent(w, r, "htmx.min.js", time.Now(), bytes.NewReader(file))
	})
}

func getEnv() {
	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file, %s", err)
	}

	MEDIAMTX_WEBRTC_URL = os.Getenv("MEDIAMTX_WEBRTC_URL")
	MEDIAMTX_HLS_URL = os.Getenv("MEDIAMTX_HLS_URL")
	MEDIAMTX_API_URL = os.Getenv("MEDIAMTX_API_URL")
	MEDIAMTX_API_PORT = os.Getenv("MEDIAMTX_API_PORT")
	MEDIAMTX_USERNAME = os.Getenv("MEDIAMTX_USERNAME")
	MEDIAMTX_PASSWORD = os.Getenv("MEDIAMTX_PASSWORD")

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
		log.Fatalf("No MediaMTX Host defined, please define the MEDIAMTX_HOST in the .env file")
	}
	if MEDIAMTX_USERNAME == "" || MEDIAMTX_PASSWORD == "" {
		log.Infof("No MEDIAMTX_USERNAME or MEDIAMTX_PASSWORD defined, no authentication will be used")
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
