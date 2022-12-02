package main

import (
	"net/http"
	"os"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jrudio/go-plex-client"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	level.Info(log).Log("msg", "Connecting", "server", serverAddress, "token", plexToken)

	conn, err := plex.New(serverAddress, plexToken)
	if err != nil {
		level.Error(log).Log("msg", "Failed to connect", "err", err)
		return
	}

	level.Info(log).Log("msg", "Successfully connected", "conn", conn)

	ctrlC := make(chan os.Signal, 1)

	onError := func(err error) {
		level.Error(log).Log("msg", "error in websocket processing", "err", err)
	}

	events := plex.NewNotificationEvents()
	events.OnPlaying(func(n plex.NotificationContainer) {
		err := onPlaying(conn, n)
		if err != nil {
			level.Error(log).Log("msg", "error handling OnPlaying event", "event", n, "err", err)
		}
	})

	// TODO - Does this automatically reconnect on websocket failure?
	conn.SubscribeToNotifications(events, ctrlC, onError)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8000", nil)
}
