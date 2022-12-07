package main

import (
	"net/http"
	"os"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grafana/plexporter/pkg/metrics"
	"github.com/grafana/plexporter/pkg/plex"
)

var (
	log = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
)

func main() {

	serverAddress := os.Getenv("PLEX_SERVER")
	if serverAddress == "" {
		level.Error(log).Log("msg", "PLEX_SERVER environment variable must be specified")
		return
	}

	plexToken := os.Getenv("PLEX_TOKEN")
	if plexToken == "" {
		level.Error(log).Log("msg", "PLEX_TOKEN environment variable must be specified")
		return
	}

	server, err := plex.NewServer(serverAddress, plexToken)
	if err != nil {
		level.Error(log).Log("msg", err)
		return
	}
	server.Listen(log)

	metrics.Register(server)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8000", nil)
}
