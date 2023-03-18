FROM alpine:3.16 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/prometheus-plex-exporter-${TARGETARCH} /prometheus-plex-exporter
ENTRYPOINT ["/prometheus-plex-exporter"]