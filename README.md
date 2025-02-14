# mediamtx-path-viewer
A way to monitor active stream paths from a MediaMTX server using the API with/without Basic Auth

This is a simple app that parses the MediaMTX API and shows a list of the paths that are currently active.  
For the API this supports the use of basic auth, if no username or password is given, it assumes that no auth is needed.

By default the app will open a web interface on http://localhost:8080/  
But you can chose with the app settings what you want it to be.

By default the app will show the HLS feed for everything other than a paths that are specifically published as a WebRTC feed

# Deployments
You can find docker builds ready for use here:  
https://hub.docker.com/r/mchauge/mediamtx-path-viewer

Or compiled builds here:  
https://github.com/McHauge/mediamtx-path-viewer/releases

# .env file params & docker variables:

## MediaMTX API
MEDIAMTX_WEBRTC_URL=https://wbrtc.example.com  
MEDIAMTX_HLS_URL=https://hls.example.com  
MEDIAMTX_API_URL=http://api.example.com  
MEDIAMTX_API_PORT=9997  
MEDIAMTX_USERNAME=(if enabled in the MediaMTX config)  
MEDIAMTX_PASSWORD=  

## App Settings
APP_PORT=8080  
APP_PATH=  
