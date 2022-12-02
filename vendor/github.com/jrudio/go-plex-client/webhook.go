package plex

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Webhook contains a webhooks information
type Webhook struct {
	Event   string `json:"event"`
	User    bool   `json:"user"`
	Owner   bool   `json:"owner"`
	Account struct {
		ID    int    `json:"id"`
		Thumb string `json:"thumb"`
		Title string `json:"title"`
	} `json:"Account"`
	Server struct {
		Title string `json:"title"`
		UUID  string `json:"uuid"`
	} `json:"Server"`
	Player struct {
		Local         bool   `json:"local"`
		PublicAddress string `json:"PublicAddress"`
		Title         string `json:"title"`
		UUID          string `json:"uuid"`
	} `json:"Player"`
	Metadata struct {
		LibrarySectionType   string `json:"librarySectionType"`
		RatingKey            string `json:"ratingKey"`
		Key                  string `json:"key"`
		ParentRatingKey      string `json:"parentRatingKey"`
		GrandparentRatingKey string `json:"grandparentRatingKey"`
		GUID                 string `json:"guid"`
		LibrarySectionID     int    `json:"librarySectionID"`
		MediaType            string `json:"type"`
		Title                string `json:"title"`
		GrandparentKey       string `json:"grandparentKey"`
		ParentKey            string `json:"parentKey"`
		GrandparentTitle     string `json:"grandparentTitle"`
		ParentTitle          string `json:"parentTitle"`
		Summary              string `json:"summary"`
		Index                int    `json:"index"`
		ParentIndex          int    `json:"parentIndex"`
		RatingCount          int    `json:"ratingCount"`
		Thumb                string `json:"thumb"`
		Art                  string `json:"art"`
		ParentThumb          string `json:"parentThumb"`
		GrandparentThumb     string `json:"grandparentThumb"`
		GrandparentArt       string `json:"grandparentArt"`
		AddedAt              int    `json:"addedAt"`
		UpdatedAt            int    `json:"updatedAt"`
	} `json:"Metadata"`
}

// WebhookEvents holds the actions for each webhook events
type WebhookEvents struct {
	events map[string]func(w Webhook)
}

// Handler listens for plex webhooks and executes the corresponding function
func (wh *WebhookEvents) Handler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(0); err != nil {
		fmt.Printf("can not read form: %v", err)
		return
	}

	var hookEvent Webhook

	payload, hasPayload := r.MultipartForm.Value["payload"]

	if hasPayload {
		if err := json.Unmarshal([]byte(payload[0]), &hookEvent); err != nil {
			fmt.Printf("can not parse json: %v", err)
			return
		}

		fn, ok := wh.events[hookEvent.Event]

		if !ok {
			fmt.Printf("unknown event name: %v\n", hookEvent.Event)
			return
		}

		fn(hookEvent)
	}
}

// newWebhookEvent attaches a function to each webhook event
func (wh *WebhookEvents) newWebhookEvent(eventName string, onEvent func(w Webhook)) error {
	switch eventName {
	case "media.play":
	case "media.pause":
	case "media.resume":
	case "media.stop":
	case "media.scrobble":
	case "media.rate":

	default:
		return errors.New("invalid event name")
	}

	wh.events[eventName] = onEvent

	return nil
}

// NewWebhook inits and returns a webhook event
func NewWebhook() *WebhookEvents {
	return &WebhookEvents{
		events: map[string]func(w Webhook){
			"media.play":     func(w Webhook) {},
			"media.pause":    func(w Webhook) {},
			"media.resume":   func(w Webhook) {},
			"media.stop":     func(w Webhook) {},
			"media.scrobble": func(w Webhook) {},
			"media.rate":     func(w Webhook) {},
		},
	}
}

// OnPlay executes when the webhook receives a play event
func (wh *WebhookEvents) OnPlay(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.play", fn)
}

// OnPause executes when the webhook receives a pause event
func (wh *WebhookEvents) OnPause(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.pause", fn)
}

// OnResume executes when the webhook receives a resume event
func (wh *WebhookEvents) OnResume(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.resume", fn)
}

// OnStop executes when the webhook receives a stop event
func (wh *WebhookEvents) OnStop(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.stop", fn)
}

// OnScrobble executes when the webhook receives a scrobble event
func (wh *WebhookEvents) OnScrobble(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.scrobble", fn)
}

// OnRate executes when the webhook receives a rate event
func (wh *WebhookEvents) OnRate(fn func(w Webhook)) error {
	return wh.newWebhookEvent("media.rate", fn)
}
