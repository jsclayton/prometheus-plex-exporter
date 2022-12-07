package plex

import (
	"strconv"
	"sync"
	"time"

	"github.com/jrudio/go-plex-client"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/grafana/plexporter/pkg/metrics"
)

type sessionState string

const (
	statePlaying   sessionState = "playing"
	stateStopped   sessionState = "stopped"
	statePaused    sessionState = "paused"
	stateBuffering sessionState = "buffering"

	mediaTypeEpisode = "episode"

	// How long metrics for sessions are kept after the last update.
	// This is used to prune prometheus metrics and keep cardinality
	// down.
	sessionTimeout = time.Minute
)

type session struct {
	session        plex.Metadata
	media          plex.Metadata
	state          sessionState
	lastUpdate     time.Time
	playStarted    time.Time
	prevPlayedTime time.Duration
}

type sessions struct {
	mtx      sync.Mutex
	sessions map[string]session
	server   *Server
}

func NewSessions(server *Server) *sessions {
	s := &sessions{
		sessions: map[string]session{},
		server:   server,
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			s.pruneOldSessions()
		}
	}()

	return s
}

func (s *sessions) pruneOldSessions() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for k, v := range s.sessions {
		if v.state == stateStopped && time.Since(v.lastUpdate) > sessionTimeout {
			delete(s.sessions, k)
		}
	}
}

func (s *sessions) Update(sessionID string, newState sessionState, newSession *plex.Metadata, media *plex.Metadata) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	session := s.sessions[sessionID]

	if newSession != nil {
		session.session = *newSession
	}

	if media != nil {
		session.media = *media
	}

	if session.state == statePlaying && newState != statePlaying {
		// If the session was playing but now is not, then flatten
		// the play time into the total.
		session.prevPlayedTime += time.Since(session.playStarted)
	}

	if session.state != statePlaying && newState == statePlaying {
		// Started playing
		session.playStarted = time.Now()
	}

	session.state = newState
	session.lastUpdate = time.Now()
	s.sessions[sessionID] = session
}

func (s *sessions) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.MetricPlayCountDesc
	ch <- metrics.MetricPlaySecondsTotalDesc
}

func (s *sessions) Collect(ch chan<- prometheus.Metric) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for id, session := range s.sessions {
		if session.playStarted.IsZero() {
			continue
		}

		title, season, episode := labels(session.media)

		ch <- metrics.Play(
			1.0,
			"plex",
			s.server.Name,
			s.server.ID,
			session.media.LibrarySectionTitle,
			session.media.LibrarySectionID.String(),
			"", // Library type?
			session.media.Type,
			title,
			season,
			episode,
			session.session.Media[0].Part[0].Decision,      // stream type
			session.session.Media[0].VideoResolution,       // stream res
			session.media.Media[0].VideoResolution,         // file res
			strconv.Itoa(session.session.Media[0].Bitrate), // bitrate
			session.session.Player.Device,                  // device
			session.session.Player.Product,                 // device type
			session.session.User.Title,
			id,
		)

		totalPlayTime := session.prevPlayedTime
		if session.state == statePlaying {
			totalPlayTime += time.Since(session.playStarted)
		}

		ch <- metrics.PlayDuration(
			float64(totalPlayTime.Seconds()),
			"plex",
			s.server.Name,
			s.server.ID,
			session.media.LibrarySectionTitle,
			session.media.LibrarySectionID.String(),
			"", // Library type?
			session.media.Type,
			title,
			season,
			episode,
			session.session.Media[0].Part[0].Decision,      // stream type
			session.session.Media[0].VideoResolution,       // stream res
			session.media.Media[0].VideoResolution,         // file res
			strconv.Itoa(session.session.Media[0].Bitrate), // bitrate
			session.session.Player.Device,                  // device
			session.session.Player.Product,                 // device type
			session.session.User.Title,
			id,
		)
	}
}

func labels(m plex.Metadata) (title, season, episodeTitle string) {
	if m.Type == mediaTypeEpisode {
		return m.GrandparentTitle, m.ParentTitle, m.Title
	}
	return m.Title, "", ""
}
