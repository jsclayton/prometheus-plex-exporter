package plex

import (
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/grafana/plexporter/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	ID      string
	Name    string
	Version string

	Token string
	URL   *url.URL

	Client *Client

	listener *plexListener

	mtx       sync.Mutex
	libraries []*Library

	lastBandwidthAt int
}

type StatisticsBandwidth struct {
	At    int   `json:"at"`
	Lan   bool  `json:"lan"`
	Bytes int64 `json:"bytes"`
}

type StatisticsResources struct {
	At             int     `json:"at"`
	HostCpuUtil    float64 `json:"hostCpuUtilization"`
	HostMemUtil    float64 `json:"hostMemoryUtilization"`
	ProcessCpuUtil float64 `json:"processCpuUtilization"`
	ProcessMemUtil float64 `json:"processMemoryUtilization"`
}

func NewServer(serverURL, token string) (*Server, error) {
	client, err := NewClient(serverURL, token)
	if err != nil {
		return nil, err
	}

	server := &Server{
		URL:   client.URL,
		Token: client.Token,

		Client:          client,
		lastBandwidthAt: int(time.Now().Unix()),
	}

	err = server.Refresh()
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			server.Refresh()
		}
	}()

	return server, nil
}

func (s *Server) Refresh() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	container := struct {
		MediaContainer struct {
			FriendlyName      string `json:"friendlyName"`
			MachineIdentifier string `json:"machineIdentifier"`
			Version           string `json:"version"`
			MediaProviders    []struct {
				Identifier string `json:"identifier"`
				Features   []struct {
					Type        string `json:"type"`
					Directories []struct {
						Identifier    string `json:"id"`
						DurationTotal int64  `json:"durationTotal"`
						StorageTotal  int64  `json:"storageTotal"`
						Title         string `json:"title"`
						Type          string `json:"type"`
					} `json:"Directory"`
				} `json:"Feature"`
			} `json:"MediaProvider"`
		} `json:"MediaContainer"`
	}{}
	err := s.Client.Get("/media/providers?includeStorage=1", &container)
	if err != nil {
		return err
	}

	s.ID = container.MediaContainer.MachineIdentifier
	s.Name = container.MediaContainer.FriendlyName
	s.Version = container.MediaContainer.Version
	s.libraries = nil
	for _, provider := range container.MediaContainer.MediaProviders {
		if provider.Identifier != "com.plexapp.plugins.library" {
			continue
		}
		for _, feature := range provider.Features {
			if feature.Type != "content" {
				continue
			}
			for _, directory := range feature.Directories {
				if !isLibraryDirectoryType(directory.Type) {
					continue
				}
				s.libraries = append(s.libraries, &Library{
					ID:            directory.Identifier,
					Name:          directory.Title,
					Type:          directory.Type,
					DurationTotal: directory.DurationTotal,
					StorageTotal:  directory.StorageTotal,
					Server:        s,
				})
			}
		}
	}

	err = s.refreshResources()
	if err != nil {
		return err
	}

	err = s.refreshBandwidth()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) refreshResources() error {
	resources := struct {
		MediaContainer struct {
			StatisticsResources []StatisticsResources `json:"StatisticsResources"`
		} `json:"MediaContainer"`
	}{}
	err := s.Client.Get("/statistics/resources?timespan=6", &resources)

	// This is a paid feature and API may not be available
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	if len(resources.MediaContainer.StatisticsResources) > 0 {
		// The last entry is the most recent
		i := len(resources.MediaContainer.StatisticsResources) - 1
		stats := resources.MediaContainer.StatisticsResources[i]

		metrics.ServerHostCpuUtilization.WithLabelValues("plex", s.Name, s.ID).Set(stats.HostCpuUtil)
		metrics.ServerHostMemUtilization.WithLabelValues("plex", s.Name, s.ID).Set(stats.HostMemUtil)
	}

	return nil
}

func (s *Server) refreshBandwidth() error {
	bandwidth := struct {
		MediaContainer struct {
			StatisticsBandwith []StatisticsBandwidth `json:"StatisticsBandwidth"`
		} `json:"MediaContainer"`
	}{}
	err := s.Client.Get("/statistics/bandwidth?timespan=6", &bandwidth)

	// This is a paid feature and API may not be available
	if err == ErrNotFound {
		return nil
	}

	if err != nil {
		return err
	}

	// Record updates newer than our last sync.  We also keep track of
	// the highest timestamp see and use that as our last sync time.
	// Sort by timestamp to ensure they are processed in order
	updates := bandwidth.MediaContainer.StatisticsBandwith

	sort.Slice(updates, func(i, j int) bool {
		return updates[i].At < updates[j].At
	})

	highest := 0
	for _, u := range updates {
		if u.At > s.lastBandwidthAt {
			metrics.MetricTransmittedBytesTotal.WithLabelValues("plex", s.Name, s.ID).Add(float64(u.Bytes))

			if u.At > highest {
				highest = u.At
			}
		}
	}

	s.lastBandwidthAt = highest

	return nil
}

func (s *Server) Library(id string) *Library {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for _, library := range s.libraries {
		if library.ID == id {
			return library
		}
	}
	return nil
}

func (s *Server) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.MetricsLibraryDurationTotalDesc
	ch <- metrics.MetricsLibraryStorageTotalDesc

	if s.listener != nil {
		s.listener.activeSessions.Describe(ch)
	}
}

func (s *Server) Collect(ch chan<- prometheus.Metric) {
	s.mtx.Lock()

	for _, library := range s.libraries {
		ch <- metrics.LibraryDuration(library.DurationTotal,
			"plex",
			library.Server.Name,
			library.Server.ID,
			library.Type,
			library.Name,
			library.ID,
		)
		ch <- metrics.LibraryStorage(library.StorageTotal,
			"plex",
			library.Server.Name,
			library.Server.ID,
			library.Type,
			library.Name,
			library.ID,
		)
	}

	// HACK: Unlock prior to asking sessions to collect since it fetches
	// 			 libraries by ID, which locks the server mutex
	s.mtx.Unlock()

	if s.listener != nil {
		s.listener.activeSessions.Collect(ch)
	}
}
