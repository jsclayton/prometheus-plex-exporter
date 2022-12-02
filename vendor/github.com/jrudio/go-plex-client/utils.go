package plex

import (
	"bytes"
	"net/http"
	"time"
)

// func (p *Plex) options(query string) (*http.Response, error) {
// 	client := p.HTTPClient

// 	req, reqErr := http.NewRequest("OPTIONS", query, nil)

// 	if reqErr != nil {
// 		return &http.Response{}, reqErr
// 	}

// 	resp, err := client.Do(req)

// 	if err != nil {
// 		return &http.Response{}, err
// 	}

// 	return resp, nil
// }

func (p *Plex) grab(query string, h headers) (*http.Response, error) {
	client := p.DownloadClient

	req, reqErr := http.NewRequest("GET", query, nil)

	if reqErr != nil {
		return &http.Response{}, reqErr
	}

	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", p.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	req.Header.Add("X-Plex-Token", p.Token)

	// optional headers
	if h.TargetClientIdentifier != "" {
		req.Header.Add("X-Plex-Target-Identifier", h.TargetClientIdentifier)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func (p *Plex) get(query string, h headers) (*http.Response, error) {
	client := p.HTTPClient

	req, reqErr := http.NewRequest("GET", query, nil)

	if reqErr != nil {
		return &http.Response{}, reqErr
	}

	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", p.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	req.Header.Add("X-Plex-Token", p.Token)

	// optional headers
	if h.TargetClientIdentifier != "" {
		req.Header.Add("X-Plex-Target-Identifier", h.TargetClientIdentifier)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func get(query string, h headers) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("GET", query, nil)

	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", h.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	if h.Token != "" {
		req.Header.Add("X-Plex-Token", h.Token)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func (p *Plex) delete(query string, h headers) (*http.Response, error) {
	client := p.HTTPClient

	req, reqErr := http.NewRequest("DELETE", query, nil)

	if reqErr != nil {
		return &http.Response{}, reqErr
	}

	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", p.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	req.Header.Add("X-Plex-Token", p.Token)

	// optional headers
	if h.TargetClientIdentifier != "" {
		req.Header.Add("X-Plex-Target-Identifier", h.TargetClientIdentifier)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func (p *Plex) post(query string, body []byte, h headers) (*http.Response, error) {
	client := p.HTTPClient

	req, err := http.NewRequest("POST", query, bytes.NewBuffer(body))

	if err != nil {
		return &http.Response{}, err
	}

	// req.Header.Set("Content-Type", applicationJson)
	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", p.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	req.Header.Add("X-Plex-Token", p.Token)
	req.Header.Add("Content-Type", h.ContentType)

	// optional headers
	if h.TargetClientIdentifier != "" {
		req.Header.Add("X-Plex-Target-Identifier", h.TargetClientIdentifier)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

// post sends a POST request and is the same as plex.post while omitting the plex token header
func post(query string, body []byte, h headers) (*http.Response, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("POST", query, bytes.NewBuffer(body))

	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", h.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	if h.Token != "" {
		req.Header.Add("X-Plex-Token", h.Token)
	}
	req.Header.Add("Content-Type", h.ContentType)

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func (p *Plex) put(query string, body []byte, h headers) (*http.Response, error) {
	client := p.HTTPClient

	req, reqErr := http.NewRequest("PUT", query, bytes.NewBuffer(body))

	if reqErr != nil {
		return &http.Response{}, reqErr
	}

	req.Header.Set("Content-Type", h.ContentType)
	req.Header.Add("Accept", h.Accept)
	req.Header.Add("X-Plex-Platform", h.Platform)
	req.Header.Add("X-Plex-Platform-Version", h.PlatformVersion)
	req.Header.Add("X-Plex-Provides", h.Provides)
	req.Header.Add("X-Plex-Client-Identifier", p.ClientIdentifier)
	req.Header.Add("X-Plex-Product", h.Product)
	req.Header.Add("X-Plex-Version", h.Version)
	req.Header.Add("X-Plex-Device", h.Device)
	// req.Header.Add("X-Plex-Container-Size", h.ContainerSize)
	// req.Header.Add("X-Plex-Container-Start", h.ContainerStart)
	req.Header.Add("X-Plex-Token", p.Token)

	// optional headers
	if h.TargetClientIdentifier != "" {
		req.Header.Add("X-Plex-Target-Identifier", h.TargetClientIdentifier)
	}

	resp, err := client.Do(req)

	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil
}

func boolToOneOrZero(input bool) string {
	if input {
		return "1"
	} else {
		return "0"
	}
}
