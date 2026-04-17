package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

type Limiter struct {
	enabled bool
	window  time.Duration
	max     int

	mu      sync.Mutex
	buckets map[string][]time.Time
}

func New(enabled bool, windowSeconds int, max int) *Limiter {
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	if max <= 0 {
		max = 120
	}
	return &Limiter{
		enabled: enabled,
		window:  time.Duration(windowSeconds) * time.Second,
		max:     max,
		buckets: map[string][]time.Time{},
	}
}

func (l *Limiter) Allow(domain, ip string, now time.Time) (bool, int) {
	if !l.enabled {
		return true, l.max
	}
	key := fmt.Sprintf("%s|%s", domain, ip)

	l.mu.Lock()
	defer l.mu.Unlock()

	hits := l.buckets[key]
	cut := now.Add(-l.window)
	pruned := hits[:0]
	for _, t := range hits {
		if t.After(cut) {
			pruned = append(pruned, t)
		}
	}
	hits = pruned
	if len(hits) >= l.max {
		l.buckets[key] = hits
		return false, 0
	}
	hits = append(hits, now)
	l.buckets[key] = hits
	remaining := l.max - len(hits)
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining
}
