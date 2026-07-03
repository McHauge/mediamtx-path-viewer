# Uses minimal Alpine base instead of the full Go toolchain image
FROM alpine:3.21

ARG VERSION
LABEL version="v${VERSION}"

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
ENV MEDIAMTX_MOQ_URL=""
ENV MEDIAMTX_PLAYBACK_URL=""

ENV APP_PORT="8080"
ENV APP_PATH=""

EXPOSE ${APP_PORT}

# Install CA certificates for HTTPS API calls
RUN apk add --no-cache ca-certificates

# Copy pre-built binary (provided by GoReleaser or downloaded manually)
COPY mediamtx-path-viewer /app/mediamtx-path-viewer
RUN chmod +x /app/mediamtx-path-viewer

# Specifies the executable command that runs when the container starts
CMD ["./mediamtx-path-viewer"]
