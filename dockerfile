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

ENV APP_PORT="8080"
ENV APP_PATH=""

EXPOSE ${APP_PORT}

# Download pre-built binary from GitHub release
RUN apk add --no-cache ca-certificates curl && \
    curl -fSL "https://github.com/McHauge/mediamtx-path-viewer/releases/download/v${VERSION}/mediamtx-path-viewer_${VERSION}_linux_amd64.tar.gz" -o /tmp/release.tar.gz && \
    tar -xzf /tmp/release.tar.gz -C /app && \
    rm /tmp/release.tar.gz && \
    apk del curl && \
    chmod +x /app/mediamtx-path-viewer

# Specifies the executable command that runs when the container starts
CMD ["./mediamtx-path-viewer"]
