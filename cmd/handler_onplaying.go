package main

import (
	"fmt"
	"time"

	"github.com/go-kit/log/level"
	"github.com/jrudio/go-plex-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	statePlaying = "playing"
	stateStopped = "stopped"
	statePaused  = "paused"
	stateBuffer  = "buffering"
)

var (
	// TODO - Add tons more labels here:  media type, library name, series name, season number, etc.
	metricPlaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plays_total",
		Help: "The total number of play counts",
	}, []string{"user", "title"})

	activeSessions = map[string]struct{}{}
)

func recordPlay(user, title string) {
	metricPlaysTotal.WithLabelValues(user, title).Inc()
}

func getSessionByID(sessions plex.CurrentSessions, sessionID string) *plex.Metadata {
	for _, session := range sessions.MediaContainer.Metadata {
		if sessionID == session.SessionKey {
			return &session
		}
	}
	return nil
}

func onPlaying(conn *plex.Plex, c plex.NotificationContainer) error {
	sessions, err := conn.GetSessions()
	if err != nil {
		return fmt.Errorf("error fetching sessions: %w", err)
	}

	for _, n := range c.PlaySessionStateNotification {
		if n.State == stateStopped {
			// When state is stopped the session is ended.
			delete(activeSessions, n.SessionKey)
			continue
		}

		session := getSessionByID(sessions, n.SessionKey)
		if session == nil {
			return fmt.Errorf("error getting session with key %s %+v", n.SessionKey, n)
		}

		metadata, err := conn.GetMetadata(n.RatingKey)
		if err != nil {
			return fmt.Errorf("error fetching metadata for key %s: %w", n.RatingKey, err)
		}

		userName := session.User.Title
		userID := session.User.ID
		mediaID := metadata.MediaContainer.Metadata[0].RatingKey
		title := metadata.MediaContainer.Metadata[0].Title
		timestamp := time.Duration(time.Millisecond) * time.Duration(n.ViewOffset)

		level.Info(log).Log("msg", "Received PlaySessionStateNotification",
			"SessionKey", n.SessionKey,
			"userName", userName,
			"userID", userID,
			"state", n.State,
			"mediaTitle", title,
			"mediaID", mediaID,
			"timestamp", timestamp)

		switch n.State {
		case statePlaying:
			if _, ok := activeSessions[n.SessionKey]; !ok {
				// New session
				activeSessions[n.SessionKey] = struct{}{}
				recordPlay(userName, title)
			}
		}
	}

	return nil
}
