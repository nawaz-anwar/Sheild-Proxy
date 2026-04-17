package challenge

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Challenge struct {
	ID         string
	Prefix     string
	Difficulty int
	ExpiresAt  time.Time
}

type stored struct {
	domain     string
	ip         string
	ua         string
	nonce      string
	difficulty int
	expiresAt  time.Time
}

type Manager struct {
	enabled    bool
	ttl        time.Duration
	difficulty int
	cookieName string

	mu   sync.Mutex
	data map[string]stored
}

func New(enabled bool, ttlSeconds int, difficulty int, cookieName string) *Manager {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	if difficulty <= 0 {
		difficulty = 3
	}
	if cookieName == "" {
		cookieName = "shield_challenge"
	}
	return &Manager{
		enabled:    enabled,
		ttl:        time.Duration(ttlSeconds) * time.Second,
		difficulty: difficulty,
		cookieName: cookieName,
		data:       map[string]stored{},
	}
}

func (m *Manager) Enabled() bool {
	return m.enabled
}

func (m *Manager) CookieName() string {
	return m.cookieName
}

func (m *Manager) Create(domain, ip, ua string, now time.Time) (Challenge, error) {
	id, err := randomHex(12)
	if err != nil {
		return Challenge{}, err
	}
	nonce, err := randomHex(16)
	if err != nil {
		return Challenge{}, err
	}
	expiresAt := now.Add(m.ttl)

	m.mu.Lock()
	m.data[id] = stored{domain: domain, ip: ip, ua: ua, nonce: nonce, difficulty: m.difficulty, expiresAt: expiresAt}
	m.mu.Unlock()

	return Challenge{ID: id, Prefix: nonce, Difficulty: m.difficulty, ExpiresAt: expiresAt}, nil
}

func (m *Manager) Verify(id, answer, domain, ip, ua string, now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.data[id]
	if !ok {
		return false
	}
	if now.After(rec.expiresAt) {
		delete(m.data, id)
		return false
	}
	if rec.domain != domain || rec.ip != ip || rec.ua != ua {
		return false
	}
	sum := sha256.Sum256([]byte(rec.nonce + answer))
	hexDigest := hex.EncodeToString(sum[:])
	if strings.HasPrefix(hexDigest, strings.Repeat("0", rec.difficulty)) {
		delete(m.data, id)
		return true
	}
	return false
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cryptographic randomness unavailable: %w", err)
	}
	return hex.EncodeToString(b), nil
}
