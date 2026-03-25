package api

import (
	"fmt"
	"sync"
	"time"
)

const (
	// SessionExpiredErrorCode is the API error code for session expiration
	SessionExpiredErrorCode = -14

	// SessionPauseDuration is how long to pause after session expiration
	SessionPauseDuration = 60 * time.Minute
)

// SessionGuard manages session state and pause periods
type SessionGuard struct {
	mu          sync.RWMutex
	pausedUntil map[string]time.Time
}

// NewSessionGuard creates a new session guard
func NewSessionGuard() *SessionGuard {
	return &SessionGuard{
		pausedUntil: make(map[string]time.Time),
	}
}

// CheckSession checks if a session is paused
func (g *SessionGuard) CheckSession(accountID string) error {
	g.mu.RLock()
	pausedUntil, exists := g.pausedUntil[accountID]
	g.mu.RUnlock()

	if !exists {
		return nil
	}

	if time.Now().Before(pausedUntil) {
		remaining := time.Until(pausedUntil)
		return fmt.Errorf("session paused for account %s (remaining: %v)", accountID, remaining)
	}

	// Pause period expired, remove it
	g.mu.Lock()
	delete(g.pausedUntil, accountID)
	g.mu.Unlock()

	return nil
}

// PauseSession marks a session as paused due to expiration
func (g *SessionGuard) PauseSession(accountID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	pauseUntil := time.Now().Add(SessionPauseDuration)
	g.pausedUntil[accountID] = pauseUntil
}

// ResumeSession resumes a paused session
func (g *SessionGuard) ResumeSession(accountID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.pausedUntil, accountID)
}

// IsPaused checks if a session is currently paused
func (g *SessionGuard) IsPaused(accountID string) bool {
	g.mu.RLock()
	pausedUntil, exists := g.pausedUntil[accountID]
	g.mu.RUnlock()

	if !exists {
		return false
	}

	return time.Now().Before(pausedUntil)
}
