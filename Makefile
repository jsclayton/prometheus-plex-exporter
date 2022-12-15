# Version number
VERSION=$(shell ./tools/image-tag | cut -d, -f 1)

GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GOPATH := $(shell go env GOPATH)

GO_OPT= -mod vendor -ldflags "-X main.Branch=$(GIT_BRANCH) -X main.Revision=$(GIT_REVISION) -X main.Version=$(VERSION)"

### Development

.PHONY: run
run:
	go run ./cmd/exporter-for-plex

### Build

.PHONY: exporter-for-plex
exporter-for-plex:
	CGO_ENABLED=0 go build $(GO_OPT) -o ./bin/$(GOOS)/exporter-for-plex-$(GOARCH) ./cmd/exporter-for-plex

.PHONY: exe
exe:
	GOOS=linux $(MAKE) $(COMPONENT)

### Docker Images

.PHONY: docker-component # Not intended to be used directly
docker-component: check-component exe
	docker build -t grafana/$(COMPONENT) --build-arg=TARGETARCH=$(GOARCH) -f ./cmd/$(COMPONENT)/Dockerfile .
	docker tag grafana/$(COMPONENT) $(COMPONENT)
	docker tag grafana/$(COMPONENT) ghcr.io/grafana/$(COMPONENT)

.PHONY: docker-exporter-for-plex
docker-exporter-for-plex:
	COMPONENT=exporter-for-plex $(MAKE) docker-component

.PHONY: docker-images
docker-images: docker-exporter-for-plex

.PHONY: check-component
check-component:
ifndef COMPONENT
	$(error COMPONENT variable was not defined)
endif
