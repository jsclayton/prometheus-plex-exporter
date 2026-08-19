FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
RUN apk --update --no-cache add git
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY pkg ./pkg
ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION
ARG GIT_REVISION
ARG GIT_BRANCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod vendor \
      -ldflags "-X main.Branch=${GIT_BRANCH} -X main.Revision=${GIT_REVISION} -X main.Version=${VERSION}" \
      -o /out/prometheus-plex-exporter ./cmd/prometheus-plex-exporter

FROM alpine:3.16 AS certs
RUN apk --update --no-cache add ca-certificates tzdata
COPY --from=build /out/prometheus-plex-exporter /prometheus-plex-exporter
ENTRYPOINT ["/prometheus-plex-exporter"]
