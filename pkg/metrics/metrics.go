package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	libraryLabels = []string{
		"server_type",  // Backend type: plex
		"server",       // Server friendly name
		"server_id",    // Server unique id
		"library_type", // movie, show, or artist ?
		"library",      // Library friendly name
		"library_id",   // Library unique id
	}

	playLabels = []string{
		"server_type",            // Backend type: plex
		"server",                 // Server friendly name
		"server_id",              // Server unique id
		"library",                // Library friendly name
		"library_id",             // Library unique id
		"library_type",           // Movies, tv_shows, music, or live_tv ?
		"media_type",             // Movies, tv_shows, music, or live_tv
		"title",                  // For tv shows this is the series title. For music this is the artist.
		"child_title",            // For tv shows this is the season title. For music this is the album title.
		"grandchild_title",       // For tv shows this is the episode title. For music this is the track title.
		"stream_type",            // DirectPlay, DirectStream, or transcode
		"stream_resolution",      // Destination resolution
		"stream_file_resolution", // Source resolution
		"stream_bitrate",         //
		"device",                 // Device friendly name
		"device_type",            //
		"user",                   // User name
		"session",
	}

	MetricsLibraryDurationTotalDesc = prometheus.NewDesc(
		"library_duration_total",
		"Total duration of a library in ms",
		libraryLabels,
		nil,
	)

	MetricsLibraryStorageTotalDesc = prometheus.NewDesc(
		"library_storage_total",
		"Total storage size of a library in Bytes",
		libraryLabels,
		nil,
	)

	MetricPlayCountDesc = prometheus.NewDesc(
		"plays_total",
		"Total play counts",
		playLabels,
		nil,
	)

	MetricPlaySecondsTotalDesc = prometheus.NewDesc(
		"play_seconds_total",
		"Total play time per session",
		playLabels,
		nil,
	)
)

func LibraryDuration(value float64,
	serverType, serverName, serverID,
	libraryType, libraryName, libraryID string,
) prometheus.Metric {

	return prometheus.MustNewConstMetric(MetricsLibraryDurationTotalDesc,
		prometheus.GaugeValue,
		value,
		serverType, serverName, serverID,
		libraryType, libraryName, libraryID,
	)
}

func LibraryStorage(value float64,
	serverType, serverName, serverID,
	libraryType, libraryName, libraryID string,
) prometheus.Metric {

	return prometheus.MustNewConstMetric(MetricsLibraryStorageTotalDesc,
		prometheus.GaugeValue,
		value,
		serverType, serverName, serverID,
		libraryType, libraryName, libraryID,
	)
}

func Play(value float64, serverType, serverName, serverID,
	library, libraryID, libraryType,
	mediaType,
	title, childTitle, grandchildTitle,
	streamType, streamResolution, streamFileResolution, streamBitrate,
	device, deviceType,
	user, session string,
) prometheus.Metric {

	return prometheus.MustNewConstMetric(MetricPlayCountDesc,
		prometheus.CounterValue,
		value,
		serverType, serverName, serverID,
		library, libraryID, libraryType,
		mediaType,
		title, childTitle, grandchildTitle,
		streamType, streamResolution, streamFileResolution, streamBitrate,
		device, deviceType,
		user, session,
	)
}

func PlayDuration(value float64, serverType, serverName, serverID,
	library, libraryID, libraryType,
	mediaType,
	title, childTitle, grandchildTitle,
	streamType, streamResolution, streamFileResolution, streamBitrate,
	device, deviceType,
	user, session string,
) prometheus.Metric {

	return prometheus.MustNewConstMetric(MetricPlaySecondsTotalDesc,
		prometheus.CounterValue,
		value,
		serverType, serverName, serverID,
		library, libraryID, libraryType,
		mediaType,
		title, childTitle, grandchildTitle,
		streamType, streamResolution, streamFileResolution, streamBitrate,
		device, deviceType,
		user, session,
	)
}
