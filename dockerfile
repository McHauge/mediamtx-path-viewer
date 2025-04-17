# Specifies a parent image
FROM golang:1.24-alpine

LABEL version="v1.3.5"

# Creates an app directory to hold your app’s source code
WORKDIR /app

# ENV Variables:
ENV MEDIAMTX_API_URL=""
ENV MEDIAMTX_API_PORT=9997
ENV MEDIAMTX_USERNAME=""
ENV MEDIAMTX_PASSWORD=""

ENV MEDIAMTX_WEBRTC_URL=""
ENV MEDIAMTX_HLS_URL=""
ENV MEDIAMTX_RTMP_URL=""
ENV MEDIAMTX_RTSP_URL=""

ENV APP_PORT="8080"
ENV APP_PATH=""

EXPOSE ${APP_PORT}

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copies everything from your root directory into /app
COPY . .

# Builds your app with optional configuration
RUN go build -v -o ./app/mediamtx-path-viewer .

# Specifies the executable command that runs when the container starts
CMD [ "./app/mediamtx-path-viewer" ]