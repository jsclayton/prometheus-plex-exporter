package plex

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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

	return client, nil
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
