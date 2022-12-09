FROM alpine:3.16 as certs
RUN apk --update add ca-certificates
ARG TARGETARCH
COPY bin/linux/exporter-for-plex-${TARGETARCH} /exporter-for-plex
ENTRYPOINT ["/exporter-for-plex"]