package main

import (
	"fmt"
	"time"

	"github.com/go-kit/log/level"
	"github.com/jrudio/go-plex-client"
)

var activeSessions *sessions

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
		if sessionState(n.State) == stateStopped {
			// When the session is stopped we can't look up the user info or media anymore.
			activeSessions.Update(n.SessionKey, sessionState(n.State), nil, nil)
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

		level.Info(log).Log("msg", "Received PlaySessionStateNotification",
			"SessionKey", n.SessionKey,
			"userName", session.User.Title,
			"userID", session.User.ID,
			"state", n.State,
			"mediaTitle", metadata.MediaContainer.Metadata[0].Title,
			"mediaID", metadata.MediaContainer.Metadata[0].RatingKey,
			"timestamp", time.Duration(time.Millisecond)*time.Duration(n.ViewOffset))

		activeSessions.Update(n.SessionKey, sessionState(n.State), &session.User, &sessions.MediaContainer.Metadata[0])
	}

	return nil
}
