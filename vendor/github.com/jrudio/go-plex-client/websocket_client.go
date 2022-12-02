package plex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// TimelineEntry ...
type TimelineEntry struct {
	Identifier    string `json:"identifier"`
	ItemID        int64  `json:"itemID"`
	MetadataState string `json:"metadataState"`
	SectionID     int64  `json:"sectionID"`
	State         int64  `json:"state"`
	Title         string `json:"title"`
	Type          int64  `json:"type"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// ActivityNotification ...
type ActivityNotification struct {
	Activity struct {
		Cancellable bool   `json:"cancellable"`
		Progress    int64  `json:"progress"`
		Subtitle    string `json:"subtitle"`
		Title       string `json:"title"`
		Type        string `json:"type"`
		UserID      int64  `json:"userID"`
		UUID        string `json:"uuid"`
	} `json:"Activity"`
	Event string `json:"event"`
	UUID  string `json:"uuid"`
}

// StatusNotification ...
type StatusNotification struct {
	Description      string `json:"description"`
	NotificationName string `json:"notificationName"`
	Title            string `json:"title"`
}

// PlaySessionStateNotification ...
type PlaySessionStateNotification struct {
	GUID             string `json:"guid"`
	Key              string `json:"key"`
	PlayQueueItemID  int64  `json:"playQueueItemID"`
	RatingKey        string `json:"ratingKey"`
	SessionKey       string `json:"sessionKey"`
	State            string `json:"state"`
	URL              string `json:"url"`
	ViewOffset       int64  `json:"viewOffset"`
	TranscodeSession string `json:"transcodeSession"`
}

// ReachabilityNotification ...
type ReachabilityNotification struct {
	Reachability bool `json:"reachability"`
}

// BackgroundProcessingQueueEventNotification ...
type BackgroundProcessingQueueEventNotification struct {
	Event   string `json:"event"`
	QueueID int64  `json:"queueID"`
}

// TranscodeSession ...
type TranscodeSession struct {
	AudioChannels        int64   `json:"audioChannels"`
	AudioCodec           string  `json:"audioCodec"`
	AudioDecision        string  `json:"audioDecision"`
	Complete             bool    `json:"complete"`
	Container            string  `json:"container"`
	Context              string  `json:"context"`
	Duration             int64   `json:"duration"`
	Key                  string  `json:"key"`
	Progress             float64 `json:"progress"`
	Protocol             string  `json:"protocol"`
	Remaining            int64   `json:"remaining"`
	SourceAudioCodec     string  `json:"sourceAudioCodec"`
	SourceVideoCodec     string  `json:"sourceVideoCodec"`
	Speed                float64 `json:"speed"`
	Throttled            bool    `json:"throttled"`
	TranscodeHwRequested bool    `json:"transcodeHwRequested"`
	VideoCodec           string  `json:"videoCodec"`
	VideoDecision        string  `json:"videoDecision"`
}

// Setting ...
type Setting struct {
	Advanced bool   `json:"advanced"`
	Default  string `json:"default"`
	Group    string `json:"group"`
	Hidden   bool   `json:"hidden"`
	ID       string `json:"id"`
	Label    string `json:"label"`
	Summary  string `json:"summary"`
	Type     string `json:"type"`
	Value    int64  `json:"value"`
}

// NotificationContainer read pms notifications
type NotificationContainer struct {
	TimelineEntry []TimelineEntry `json:"TimelineEntry"`

	ActivityNotification []ActivityNotification `json:"ActivityNotification"`

	StatusNotification []StatusNotification `json:"StatusNotification"`

	PlaySessionStateNotification []PlaySessionStateNotification `json:"PlaySessionStateNotification"`

	ReachabilityNotification []ReachabilityNotification `json:"ReachabilityNotification"`

	BackgroundProcessingQueueEventNotification []BackgroundProcessingQueueEventNotification `json:"BackgroundProcessingQueueEventNotification"`

	TranscodeSession []TranscodeSession `json:"TranscodeSession"`

	Setting []Setting `json:"Setting"`

	Size int64 `json:"size"`
	// Type can be one of:
	// playing,
	// reachability,
	// transcode.end,
	// preference,
	// update.statechange,
	// activity,
	// backgroundProcessingQueue,
	// transcodeSession.update
	// transcodeSession.end
	Type string `json:"type"`
}

// WebsocketNotification websocket payload of notifications from a plex media server
type WebsocketNotification struct {
	NotificationContainer `json:"NotificationContainer"`
}

// NotificationEvents hold callbacks that correspond to notifications
type NotificationEvents struct {
	events map[string]func(n NotificationContainer)
}

// NewNotificationEvents initializes the event callbacks
func NewNotificationEvents() *NotificationEvents {
	return &NotificationEvents{
		events: map[string]func(n NotificationContainer){
			"playing":                   func(n NotificationContainer) {},
			"reachability":              func(n NotificationContainer) {},
			"transcode.end":             func(n NotificationContainer) {},
			"transcodeSession.end":      func(n NotificationContainer) {},
			"transcodeSession.update":   func(n NotificationContainer) {},
			"preference":                func(n NotificationContainer) {},
			"update.statechange":        func(n NotificationContainer) {},
			"activity":                  func(n NotificationContainer) {},
			"backgroundProcessingQueue": func(n NotificationContainer) {},
		},
	}
}

// OnPlaying shows state information (resume, stop, pause) on a user consuming media in plex
func (e *NotificationEvents) OnPlaying(fn func(n NotificationContainer)) {
	e.events["playing"] = fn
}

// OnTranscodeUpdate shows transcode information when a transcoding stream changes parameters
func (e *NotificationEvents) OnTranscodeUpdate(fn func(n NotificationContainer)) {
	e.events["transcodeSession.update"] = fn
}

// SubscribeToNotifications connects to your server via websockets listening for events
func (p *Plex) SubscribeToNotifications(events *NotificationEvents, interrupt <-chan os.Signal, fn func(error)) {
	plexURL, err := url.Parse(p.URL)

	if err != nil {
		fn(err)
		return
	}

	websocketURL := url.URL{Scheme: "ws", Host: plexURL.Host, Path: "/:/websockets/notifications"}

	headers := http.Header{
		"X-Plex-Token": []string{p.Token},
	}

	c, _, err := websocket.DefaultDialer.Dial(websocketURL.String(), headers)

	if err != nil {
		fn(err)
		return
	}

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)

		for {
			_, message, err := c.ReadMessage()

			if err != nil {
				fmt.Println("read:", err)
				fn(err)
				return
			}

			// fmt.Printf("\t%s\n", string(message))

			var notif WebsocketNotification

			if err := json.Unmarshal(message, &notif); err != nil {
				fmt.Printf("convert message to json failed: %v\n", err)
				continue
			}

			// fmt.Println(notif.Type)
			fn, ok := events.events[notif.Type]

			if !ok {
				fmt.Printf("unknown websocket event name: %v\n", notif.Type)
				continue
			}

			fn(notif.NotificationContainer)
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))

				if err != nil {
					fn(err)
				}
			case <-interrupt:
				fmt.Println("interrupt")
				// To cleanly close a connection, a client should send a close
				// frame and wait for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

				if err != nil {
					fmt.Println("write close:", err)
					fn(err)
				}

				select {
				case <-done:
				case <-time.After(time.Second):
					fmt.Println("closing websocket...")
					c.Close()
				}
				return
			}
		}
	}()
}
