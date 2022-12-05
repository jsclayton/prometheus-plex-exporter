package main

import (
	"sync"
	"time"

	"github.com/jrudio/go-plex-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type sessionState string

const (
	statePlaying   sessionState = "playing"
	stateStopped   sessionState = "stopped"
	statePaused    sessionState = "paused"
	stateBuffering sessionState = "buffering"

	sessionTimeout = time.Minute
)

var (
	metricPlaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plays_total",
		Help: "The total number of play counts",
	}, []string{"library_section", "media_type", "grandparent_title", "parent_title", "title", "user"})

	metricPlaySecondsTotalDesc = prometheus.NewDesc(
		"play_seconds_total",
		"Total play time per session",
		[]string{"library_section", "media_type", "grandparent_title", "parent_title", "title", "user", "session"},
		nil,
	)

	activeSessions = NewSessions()
)

type session struct {
	user           plex.User
	media          plex.Metadata
	state          sessionState
	lastUpdate     time.Time
	playStarted    time.Time
	prevPlayedTime time.Duration
}

type sessions struct {
	mtx      sync.Mutex
	sessions map[string]session
}

func NewSessions() *sessions {
	s := &sessions{
		sessions: map[string]session{},
	}

	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			s.pruneOldSessions()
		}
	}()

	prometheus.MustRegister(s)
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

func (s *sessions) Update(sessionID string, newState sessionState, user *plex.User, media *plex.Metadata) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	session, ok := s.sessions[sessionID]

	// If new session playing something, then record a new play
	if !ok && newState == statePlaying && user != nil && media != nil {
		metricPlaysTotal.WithLabelValues(
			media.LibrarySectionTitle,
			media.Type,
			media.GrandparentTitle,
			media.ParentTitle,
			media.Title,
			user.Title,
		).Inc()
	}

	if user != nil {
		session.user = *user
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
	ch <- metricPlaySecondsTotalDesc
}

func (s *sessions) Collect(ch chan<- prometheus.Metric) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for id, session := range s.sessions {

		totalPlayTime := session.prevPlayedTime

		if session.state == statePlaying {
			totalPlayTime += time.Since(session.playStarted)
		}

		ch <- prometheus.MustNewConstMetric(metricPlaySecondsTotalDesc,
			prometheus.CounterValue,
			float64(totalPlayTime.Seconds()),
			session.media.LibrarySectionTitle,
			session.media.Type,
			session.media.GrandparentTitle, // for tv shows this is the series
			session.media.ParentTitle,      // for tv shows this is the season
			session.media.Title,            // for tv shows this is the episode title
			session.user.Title,
			id,
		)
	}
}
