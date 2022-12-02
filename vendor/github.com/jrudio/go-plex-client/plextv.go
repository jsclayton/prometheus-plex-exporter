package plex

// I'll slowly migrate plex.tv related functions to this file

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ErrorResponse contains a code and an error message
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PinResponse holds information to successfully check a pin when linking an account
type PinResponse struct {
	ID               int             `json:"id"`
	Code             string          `json:"code"`
	ClientIdentifier string          `json:"clientIdentifier"`
	CreatedAt        string          `json:"createdAt"`
	ExpiresAt        string          `json:"expiresAt"`
	ExpiresIn        json.Number     `json:"expiresIn"`
	AuthToken        string          `json:"authToken"`
	Errors           []ErrorResponse `json:"errors"`
	Trusted          bool            `json:"trusted"`
	Location         struct {
		Code         string `json:"code"`
		Country      string `json:"country"`
		City         string `json:"city"`
		Subdivisions string `json:"subdivisions"`
		Coordinates  string `json:"coordinates"`
	}
}

// RequestPIN will retrieve a code (valid for 15 minutes) from plex.tv to link an app to your plex account
func RequestPIN(requestHeaders headers) (PinResponse, error) {
	endpoint := "/api/v2/pins.json"

	// POST request and returns a 201 status code
	// response body returns json
	//
	// {
	// 		id: 123456,
	// 		code: "ABCD",
	// 		clientIdentifier: "goplexclient",
	// 		expiresAt: 15463757,
	// 		authToken: null
	// }
	var pinInformation PinResponse

	if requestHeaders.ClientIdentifier == "" {
		requestHeaders = defaultHeaders()
	}

	resp, err := post(plexURL+endpoint, nil, requestHeaders)

	if err != nil {
		return pinInformation, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return pinInformation, errors.New(resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&pinInformation); err != nil {
		return pinInformation, err
	}

	return pinInformation, nil
}

// CheckPIN will return information related to the pin such as the auth token if your code has been approved.
// will return an error if code expired or still not linked
// clientIdentifier must be the same when requesting a pin
func CheckPIN(id int, clientIdentifier string) (PinResponse, error) {
	endpoint := "/api/v2/pins/"

	endpoint = endpoint + strconv.Itoa(id) + ".json"

	headers := defaultHeaders()

	if clientIdentifier != "" {
		headers.ClientIdentifier = clientIdentifier
	}

	resp, err := get(plexURL+endpoint, headers)

	if err != nil {
		return PinResponse{}, err
	}

	defer resp.Body.Close()

	var pinInformation PinResponse

	if err := json.NewDecoder(resp.Body).Decode(&pinInformation); err != nil {
		return pinInformation, err
	}

	// code doesn't exist or expired
	if len(pinInformation.Errors) > 0 {
		return pinInformation, errors.New(pinInformation.Errors[0].Message)
	}

	// we are not authorized yet
	if pinInformation.AuthToken == "" {
		return pinInformation, errors.New(ErrorPINNotAuthorized)
	}

	// we are authorized! Yay!
	return pinInformation, nil
}

// LinkAccount allows you to authorize an app via a 4 character pin. returns nil on success
func (p Plex) LinkAccount(code string) error {
	endpoint := "/api/v2/pins/link.json"

	body := url.Values{
		"code": []string{code},
	}

	headers := p.Headers

	headers.ContentType = "application/x-www-form-urlencoded"

	// PUT request with 'code: <4-character-pin>' in the body
	resp, err := p.put(plexURL+endpoint, []byte(body.Encode()), headers)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// type linkAccountResponse struct {

	// }

	// var

	// json.NewDecoder(resp.Body).Decode()

	// should return 204 for success
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf(ErrorLinkAccount, resp.Status)
	}

	return nil
}

type webhookErr struct {
	Err []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"errors"`
}

func (w webhookErr) Error() string {
	if len(w.Err) == 0 {
		return ""
	}

	return w.Err[0].Message
}

// GetWebhooks fetches all webhooks - requires plex pass
func (p Plex) GetWebhooks() ([]string, error) {
	type Hooks struct {
		URL string `json:"url"`
	}

	var webhooks []string

	endpoint := "/api/v2/user/webhooks"

	resp, err := p.get(plexURL+endpoint, p.Headers)

	if err != nil {
		return webhooks, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		var webhookErr webhookErr

		if err := json.NewDecoder(resp.Body).Decode(&webhookErr); err != nil {
			return webhooks, err
		}

		return webhooks, fmt.Errorf(ErrorWebhook, webhookErr.Error())
	} else if resp.StatusCode != http.StatusOK {
		return webhooks, fmt.Errorf(ErrorWebhook, resp.Status)
	}

	var hook []Hooks

	if err := json.NewDecoder(resp.Body).Decode(&hook); err != nil {
		return webhooks, err
	}

	for _, h := range hook {
		webhooks = append(webhooks, h.URL)
	}

	return webhooks, nil
}

// AddWebhook creates a new webhook for your plex server to send metadata - requires plex pass
func (p Plex) AddWebhook(webhook string) error {
	// get current webhooks and append ours to it
	currentWebhooks, err := p.GetWebhooks()

	if err != nil {
		return err
	}

	currentWebhooks = append(currentWebhooks, webhook)

	return p.SetWebhooks(currentWebhooks)
}

// SetWebhooks will set your webhooks to whatever you pass as an argument
// webhooks with a length of 0 will remove all webhooks
func (p Plex) SetWebhooks(webhooks []string) error {
	endpoint := "/api/v2/user/webhooks"

	body := url.Values{}

	if len(webhooks) == 0 {
		body.Add("urls[]", "")
	}

	for _, hook := range webhooks {
		body.Add("urls[]", hook)
	}

	headers := p.Headers

	headers.ContentType = "application/x-www-form-urlencoded"

	resp, err := p.post(plexURL+endpoint, []byte(body.Encode()), headers)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.New(ErrorFailedToSetWebhook)
	}

	return nil
}

// MyAccount gets account info (i.e. plex pass, servers, username, etc) from plex tv
func (p Plex) MyAccount() (UserPlexTV, error) {
	endpoint := "/users/account"

	var account UserPlexTV

	resp, err := p.get(plexURL+endpoint, p.Headers)

	if err != nil {
		return account, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return account, errors.New(ErrorInvalidToken)
	} else if resp.StatusCode != http.StatusOK {
		return account, errors.New(resp.Status)
	}

	if err := xml.NewDecoder(resp.Body).Decode(&account); err != nil {
		return account, err
	}

	return account, err
}
