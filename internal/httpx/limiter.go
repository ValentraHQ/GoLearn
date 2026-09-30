package httpx

import (
	"sync"
	"time"
)

// Limiter is a keyed token-bucket rate limiter.
type Limiter struct {
	mu      sync.Mutex
	rate    float64 // tokens per second
	burst   float64
	buckets map[string]*bucket
	last    time.Time
	now     func() time.Time
}

type bucket struct {
	tokens float64
	at     time.Time
}

func NewLimiter(perMinute int, burst int) *Limiter {
	return &Limiter{rate: float64(perMinute) / 60, burst: float64(burst), buckets: map[string]*bucket{}, now: time.Now}
}

// Allow consumes one token for key; it returns false and a retry hint when empty.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	n := l.now()
	if n.Sub(l.last) > 5*time.Minute {
		for k, b := range l.buckets {
			if n.Sub(b.at) > 10*time.Minute {
				delete(l.buckets, k)
			}
		}
		l.last = n
	}
	b := l.buckets[key]
	if b == nil {
		b = &bucket{tokens: l.burst, at: n}
		l.buckets[key] = b
	}
	b.tokens += n.Sub(b.at).Seconds() * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.at = n
	if b.tokens < 1 {
		return false, time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
	}
	b.tokens--
	return true, 0
}
