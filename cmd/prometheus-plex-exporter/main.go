package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/grafana/plexporter/pkg/metrics"
	"github.com/grafana/plexporter/pkg/plex"
)

const (
	MetricsServerAddr = ":9000"
)

var (
	log = kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	serverAddress := os.Getenv("PLEX_SERVER")
	if serverAddress == "" {
		level.Error(log).Log("msg", "PLEX_SERVER environment variable must be specified")
		os.Exit(1)
	}

	plexToken := os.Getenv("PLEX_TOKEN")
	if plexToken == "" {
		level.Error(log).Log("msg", "PLEX_TOKEN environment variable must be specified")
		os.Exit(1)
	}

	server, err := plex.NewServer(serverAddress, plexToken)
	if err != nil {
		level.Error(log).Log("msg", "cannot initialize connection to plex server", "error", err)
		os.Exit(1)
	}

	metrics.Register(server)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsServer := http.Server{
		Addr:         MetricsServerAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		level.Info(log).Log("msg", "starting metrics server on "+MetricsServerAddr)
		err = metricsServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			level.Error(log).Log("msg", "cannot start metrics server", "error", err)
		}
	}()

	exitCode := 0
	err = server.Listen(ctx, log)
	if err != nil {
		level.Error(log).Log("msg", "cannot listen to plex server events", "error", err)
		exitCode = 1
	}

	level.Debug(log).Log("msg", "shutting down metrics server")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		level.Error(log).Log("msg", "cannot gracefully shutdown metrics server", "error", err)
	}

	os.Exit(exitCode)
}
