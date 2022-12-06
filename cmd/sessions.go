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

	mediaTypeEpisode = "episode"

	sessionTimeout = time.Minute
)

var (
	metricPlaysTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "plays_total",
		Help: "The total number of play counts",
	}, []string{"server", "library_section", "media_type", "title", "season", "episode_title", "user"})

	metricPlaySecondsTotalDesc = prometheus.NewDesc(
		"play_seconds_total",
		"Total play time per session",
		[]string{"server", "library_section", "media_type", "title", "season", "episode_title", "user", "session"},
		nil,
	)
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
	mtx        sync.Mutex
	sessions   map[string]session
	serverName string
}

func NewSessions(serverName string) *sessions {
	s := &sessions{
		sessions:   map[string]session{},
		serverName: serverName,
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

	session := s.sessions[sessionID]

	// If session playing something for the first time, then record a new play
	if newState == statePlaying && session.playStarted.IsZero() && user != nil && media != nil {
		title, season, episode := labels(*media)
		metricPlaysTotal.WithLabelValues(
			s.serverName,
			media.LibrarySectionTitle,
			media.Type,
			title,
			season,
			episode,
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

		title, season, episode := labels(session.media)

		ch <- prometheus.MustNewConstMetric(metricPlaySecondsTotalDesc,
			prometheus.CounterValue,
			float64(totalPlayTime.Seconds()),
			s.serverName,
			session.media.LibrarySectionTitle,
			session.media.Type,
			title,
			season,
			episode,
			session.user.Title,
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
