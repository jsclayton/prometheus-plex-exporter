package plex

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/grafana/plexporter/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Client struct {
	Identifier string
	Name       string
	Token      string
	URL        *url.URL
	Version    string

	httpClient http.Client
}

func NewClient(serverURL, token string) (*Client, error) {
	parsed, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Token:      token,
		URL:        parsed,
		httpClient: http.Client{},
	}

	rootContainer := struct {
		MediaContainer struct {
			FriendlyName      string `json:"friendlyName"`
			MachineIdentifier string `json:"machineIdentifier"`
			Version           string `json:"version"`
		} `json:"MediaContainer"`
	}{}
	err = client.Get("/", &rootContainer)
	if err != nil {
		return nil, err
	}

	client.Identifier = rootContainer.MediaContainer.MachineIdentifier
	client.Name = rootContainer.MediaContainer.FriendlyName
	client.Version = rootContainer.MediaContainer.Version

	prometheus.MustRegister(client)

	return client, nil
}

func isInterestingDirectoryType(directoryType string) bool {
	switch directoryType {
	case
		"movie",
		"show",
		"artist":
		return true
	}
	return false
}

func (c *Client) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.MetricsLibraryDurationTotalDesc
	ch <- metrics.MetricsLibraryStorageTotalDesc
}

func (c *Client) Collect(ch chan<- prometheus.Metric) {
	container := struct {
		MediaContainer struct {
			MediaProviders []struct {
				Identifier string `json:"identifier"`
				Features   []struct {
					Type        string `json:"type"`
					Directories []struct {
						Identifier    string  `json:"id"`
						DurationTotal float64 `json:"durationTotal"`
						StorageTotal  float64 `json:"storageTotal"`
						Title         string  `json:"title"`
						Type          string  `json:"type"`
					} `json:"Directory"`
				} `json:"Feature"`
			} `json:"MediaProvider"`
		} `json:"MediaContainer"`
	}{}
	err := c.Get("/media/providers?includeStorage=1", &container)
	if err != nil {
		return
	}

	for _, provider := range container.MediaContainer.MediaProviders {
		if provider.Identifier != "com.plexapp.plugins.library" {
			continue
		}
		for _, feature := range provider.Features {
			if feature.Type != "content" {
				continue
			}
			for _, directory := range feature.Directories {
				if !isInterestingDirectoryType(directory.Type) {
					continue
				}
				ch <- metrics.LibraryDuration(directory.DurationTotal,
					"plex",
					c.Name,
					c.Identifier,
					directory.Type,
					directory.Title,
					directory.Identifier,
				)
				ch <- metrics.LibraryStorage(directory.StorageTotal,
					"plex",
					c.Name,
					c.Identifier,
					directory.Type,
					directory.Title,
					directory.Identifier,
				)
			}
		}
	}
}

func (c *Client) NewRequest(method, path string) (*http.Request, error) {
	requestPath, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	reqURL := c.URL.ResolveReference(requestPath)
	req, err := http.NewRequest(method, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Plex-Token", c.Token)

	return req, nil
}

func (c *Client) Do(request *http.Request, data any) error {
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}

func (c *Client) Get(path string, data any) error {
	req, err := c.NewRequest("GET", path)
	if err != nil {
		return err
	}

	return c.Do(req, &data)
}
