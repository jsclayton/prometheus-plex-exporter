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
	mtx                            sync.Mutex
	sessions                       map[string]session
	server                         *Server
	totalEstimatedTransmittedKBits float64
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

	ss := s.sessions[sessionID]

	if newSession != nil {
		ss.session = *newSession
	}

	if media != nil {
		ss.media = *media
	}

	if ss.state == statePlaying && newState != statePlaying {
		// If the session was playing but now is not, then flatten
		// the play time into the total.
		ss.prevPlayedTime += time.Since(ss.playStarted)
		s.totalEstimatedTransmittedKBits += time.Since(ss.playStarted).Seconds() * float64(ss.session.Media[0].Bitrate)
	}

	if ss.state != statePlaying && newState == statePlaying {
		// Started playing
		ss.playStarted = time.Now()
	}

	ss.state = newState
	ss.lastUpdate = time.Now()
	s.sessions[sessionID] = ss
}

func (s *sessions) extrapolatedTransmittedBytes() float64 {

	total := s.totalEstimatedTransmittedKBits

	for _, ss := range s.sessions {
		if ss.state == statePlaying {
			total += time.Since(ss.playStarted).Seconds() * float64(ss.session.Media[0].Bitrate)
		}
	}

	return total * 128.0 // Kbits -> Bytes, 1024 / 8
}

func (s *sessions) Describe(ch chan<- *prometheus.Desc) {
	ch <- metrics.MetricPlayCountDesc
	ch <- metrics.MetricPlaySecondsTotalDesc

	ch <- metrics.MetricEstimatedTransmittedBytesTotal
}

func (s *sessions) Collect(ch chan<- prometheus.Metric) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	for id, session := range s.sessions {
		if session.playStarted.IsZero() {
			continue
		}

		title, season, episode := labels(session.media)
		library := s.server.Library(session.media.LibrarySectionID.String())
		if library == nil {
			continue
		}

		ch <- metrics.Play(
			1.0,
			"plex",
			s.server.Name,
			s.server.ID,
			library.Name,
			library.ID,
			library.Type,
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
			library.Name,
			library.ID,
			library.Type,
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

	ch <- prometheus.MustNewConstMetric(metrics.MetricEstimatedTransmittedBytesTotal, prometheus.CounterValue, s.extrapolatedTransmittedBytes(), "plex", s.server.Name,
		s.server.ID)
}

func labels(m plex.Metadata) (title, season, episodeTitle string) {
	if m.Type == mediaTypeEpisode {
		return m.GrandparentTitle, m.ParentTitle, m.Title
	}
	return m.Title, "", ""
}
