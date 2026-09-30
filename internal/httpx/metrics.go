package httpx

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Metrics is a tiny Prometheus text-format exporter (counters + latency sums).
type Metrics struct {
	mu       sync.Mutex
	requests map[string]uint64
	durSum   map[string]float64
	custom   map[string]uint64
}

func NewMetrics() *Metrics {
	return &Metrics{requests: map[string]uint64{}, durSum: map[string]float64{}, custom: map[string]uint64{}}
}

func (m *Metrics) Observe(method, route string, status int, d time.Duration) {
	key := fmt.Sprintf(`method="%s",route="%s",code="%d"`, method, route, status)
	m.mu.Lock()
	m.requests[key]++
	m.durSum[key] += d.Seconds()
	m.mu.Unlock()
}

// Inc increments a named counter with a pre-rendered label string, e.g. `result="ok"`.
func (m *Metrics) Inc(name, labels string) {
	m.mu.Lock()
	m.custom[name+"{"+labels+"}"]++
	m.mu.Unlock()
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		var b strings.Builder
		b.WriteString("# TYPE golearn_http_requests_total counter\n")
		for _, k := range sortedKeys(m.requests) {
			fmt.Fprintf(&b, "golearn_http_requests_total{%s} %d\n", k, m.requests[k])
		}
		b.WriteString("# TYPE golearn_http_request_duration_seconds_sum counter\n")
		for _, k := range sortedKeys(m.requests) {
			fmt.Fprintf(&b, "golearn_http_request_duration_seconds_sum{%s} %g\n", k, m.durSum[k])
		}
		for _, k := range sortedKeys(m.custom) {
			fmt.Fprintf(&b, "%s %d\n", k, m.custom[k])
		}
		_, _ = w.Write([]byte(b.String()))
	})
}

func sortedKeys(m map[string]uint64) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}
