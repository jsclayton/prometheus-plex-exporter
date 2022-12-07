package plex

import (
	"fmt"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jrudio/go-plex-client"
)

type plexListener struct {
	server         *Server
	conn           *plex.Plex
	activeSessions *sessions
	log            log.Logger
}

func (s *Server) Listen(log log.Logger) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	conn, err := plex.New(s.URL.String(), s.Token)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", s.URL.String(), err)
	}

	// TODO: Gracefully close any previous listener
	s.listener = &plexListener{
		server:         s,
		conn:           conn,
		activeSessions: NewSessions(s),
		log:            log,
	}

	ctrlC := make(chan os.Signal, 1)

	onError := func(err error) {
		level.Error(log).Log("msg", "error in websocket processing", "err", err)
	}

	events := plex.NewNotificationEvents()
	events.OnPlaying(s.listener.onPlayingHandler)

	// TODO - Does this automatically reconnect on websocket failure?
	conn.SubscribeToNotifications(events, ctrlC, onError)

	level.Info(log).Log("msg", "Successfully connected", "machineID", s.ID, "server", s.Name)

	return nil
}

func getSessionByID(sessions plex.CurrentSessions, sessionID string) *plex.Metadata {
	for _, session := range sessions.MediaContainer.Metadata {
		if sessionID == session.SessionKey {
			return &session
		}
	}
	return nil
}

func (l *plexListener) onPlayingHandler(c plex.NotificationContainer) {
	err := l.onPlaying(c)
	if err != nil {
		level.Error(l.log).Log("msg", "error handling OnPlaying event", "event", c, "err", err)
	}
}

func (l *plexListener) onPlaying(c plex.NotificationContainer) error {
	sessions, err := l.conn.GetSessions()
	if err != nil {
		return fmt.Errorf("error fetching sessions: %w", err)
	}

	for _, n := range c.PlaySessionStateNotification {
		if sessionState(n.State) == stateStopped {
			// When the session is stopped we can't look up the user info or media anymore.
			l.activeSessions.Update(n.SessionKey, sessionState(n.State), nil, nil)
			continue
		}

		session := getSessionByID(sessions, n.SessionKey)
		if session == nil {
			return fmt.Errorf("error getting session with key %s %+v", n.SessionKey, n)
		}

		metadata, err := l.conn.GetMetadata(n.RatingKey)
		if err != nil {
			return fmt.Errorf("error fetching metadata for key %s: %w", n.RatingKey, err)
		}

		level.Info(l.log).Log("msg", "Received PlaySessionStateNotification",
			"SessionKey", n.SessionKey,
			"userName", session.User.Title,
			"userID", session.User.ID,
			"state", n.State,
			"mediaTitle", metadata.MediaContainer.Metadata[0].Title,
			"mediaID", metadata.MediaContainer.Metadata[0].RatingKey,
			"timestamp", time.Duration(time.Millisecond)*time.Duration(n.ViewOffset))

		l.activeSessions.Update(n.SessionKey, sessionState(n.State), session, &metadata.MediaContainer.Metadata[0])
	}

	return nil
}
