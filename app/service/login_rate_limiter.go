package service

import (
	"sync"
	"time"
)

type LoginRateLimiter struct {
	maxFails   int
	lockoutDur time.Duration

	mu     sync.Mutex
	fails  map[string]*failRecord
	ticker *time.Ticker
}

type failRecord struct {
	count     int
	lockUntil time.Time
}

func NewLoginRateLimiter(maxFails int, lockout time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		maxFails:   maxFails,
		lockoutDur: lockout,
		fails:      make(map[string]*failRecord),
	}
}

func (l *LoginRateLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.fails[key]
	if !ok {
		return true, 0
	}
	if time.Now().Before(r.lockUntil) {
		return false, time.Until(r.lockUntil)
	}
	if !r.lockUntil.IsZero() && time.Now().After(r.lockUntil) {
		delete(l.fails, key)
	}
	return true, 0
}

func (l *LoginRateLimiter) RegisterFailure(key string) (int, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.fails[key]
	if !ok {
		r = &failRecord{}
		l.fails[key] = r
	}
	r.count++
	if r.count >= l.maxFails {
		r.lockUntil = time.Now().Add(l.lockoutDur)
		return r.count, l.lockoutDur
	}
	return r.count, 0
}

func (l *LoginRateLimiter) RegisterSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}

func (l *LoginRateLimiter) Start(interval time.Duration) {
	l.ticker = time.NewTicker(interval)
	go func() {
		for range l.ticker.C {
			l.mu.Lock()
			now := time.Now()
			for k, r := range l.fails {
				if !r.lockUntil.IsZero() && now.After(r.lockUntil) {
					delete(l.fails, k)
				}
			}
			l.mu.Unlock()
		}
	}()
}

func (l *LoginRateLimiter) Stop() {
	if l.ticker != nil {
		l.ticker.Stop()
	}
}
