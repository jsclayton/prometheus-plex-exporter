package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MetricPlaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plays_total",
		Help: "The total number of play counts",
	}, []string{"server", "library_section", "media_type", "title", "season", "episode_title", "user"})

	MetricPlaySecondsTotalDesc = prometheus.NewDesc(
		"play_seconds_total",
		"Total play time per session",
		[]string{"server", "library_section", "media_type", "title", "season", "episode_title", "user", "session"},
		nil,
	)
)

func Play(server, librarySection, mediaType, title, season, episodeTitle, user string) {
	MetricPlaysTotal.WithLabelValues(
		server,
		librarySection,
		mediaType,
		title,
		season,
		episodeTitle,
		user).Inc()
}

func PlayDuration(value float64, server, librarySection, mediaType, title, season, episodeTitle, user, session string) prometheus.Metric {
	return prometheus.MustNewConstMetric(MetricPlaySecondsTotalDesc,
		prometheus.CounterValue,
		value,
		server,
		librarySection,
		mediaType,
		title,
		season,
		episodeTitle,
		user,
		session,
	)
}
