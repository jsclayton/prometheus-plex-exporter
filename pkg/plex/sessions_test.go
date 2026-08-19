package plex

import (
	"context"
	"testing"

	"github.com/jrudio/go-plex-client"
)

func newTestSessions() *sessions {
	return NewSessions(context.Background(), &Server{ID: "test", Name: "test"}, nil)
}

func metadataWithBitrate(bitrate int) *plex.Metadata {
	return &plex.Metadata{
		Media: []plex.Media{
			{Bitrate: bitrate},
		},
	}
}

func TestReconcileActive_StopsMissingSessions(t *testing.T) {
	s := newTestSessions()

	s.Update("abc123", statePlaying, metadataWithBitrate(1000), metadataWithBitrate(1000))

	if got := s.sessions["abc123"].state; got != statePlaying {
		t.Fatalf("expected session to be playing after Update, got %v", got)
	}

	s.ReconcileActive(map[string]struct{}{})

	if got := s.sessions["abc123"].state; got != stateStopped {
		t.Fatalf("expected session to be stopped after reconcile, got %v", got)
	}
}

func TestReconcileActive_LeavesLiveSessionsAlone(t *testing.T) {
	s := newTestSessions()

	s.Update("abc123", statePlaying, metadataWithBitrate(1000), metadataWithBitrate(1000))

	s.ReconcileActive(map[string]struct{}{"abc123": {}})

	if got := s.sessions["abc123"].state; got != statePlaying {
		t.Fatalf("expected session to remain playing, got %v", got)
	}
}

func TestReconcileActive_AlreadyStoppedUntouched(t *testing.T) {
	s := newTestSessions()

	s.Update("abc123", statePlaying, metadataWithBitrate(1000), metadataWithBitrate(1000))
	s.Update("abc123", stateStopped, nil, nil)

	before := s.totalEstimatedTransmittedKBits

	s.ReconcileActive(map[string]struct{}{})

	if got := s.sessions["abc123"].state; got != stateStopped {
		t.Fatalf("expected session to remain stopped, got %v", got)
	}
	if s.totalEstimatedTransmittedKBits != before {
		t.Fatalf("expected totalEstimatedTransmittedKBits to be unchanged, got %v want %v", s.totalEstimatedTransmittedKBits, before)
	}
}
