// Package monitoring records aggregate edge metrics without request payloads or tokens.
package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

var bounds = []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10}

type Metrics struct {
	mu      sync.Mutex
	count   uint64
	errors  uint64
	seconds float64
	buckets []uint64
}

func New() *Metrics { return &Metrics{buckets: make([]uint64, len(bounds))} }
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "/metrics" || route == "/healthz" || route == "/readyz" {
			return
		}
		if route == "" {
			route = "unmatched"
		}
		elapsed := time.Since(started).Seconds()
		m.mu.Lock()
		m.count++
		if c.Writer.Status() >= 400 {
			m.errors++
		}
		m.seconds += elapsed
		for i, bound := range bounds {
			if elapsed <= bound {
				m.buckets[i]++
			}
		}
		m.mu.Unlock()
		// Route templates omit user IDs; bodies, queries, headers, and tokens are never logged here.
		record, _ := json.Marshal(map[string]any{"event": "http_request", "route": route, "status": c.Writer.Status(), "duration_ms": elapsed * 1000})
		log.Print(string(record))
	}
}
func (m *Metrics) Handler(c *gin.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.Status(http.StatusOK)
	fmt.Fprintf(c.Writer, "# HELP http_requests_total API requests excluding probes and metrics.\n# TYPE http_requests_total counter\nhttp_requests_total %d\n", m.count)
	fmt.Fprintf(c.Writer, "# HELP http_errors_total Requests with HTTP status 400 or higher.\n# TYPE http_errors_total counter\nhttp_errors_total %d\n", m.errors)
	fmt.Fprint(c.Writer, "# HELP http_request_duration_seconds API request latency.\n# TYPE http_request_duration_seconds histogram\n")
	for i, bound := range bounds {
		fmt.Fprintf(c.Writer, "http_request_duration_seconds_bucket{le=\"%g\"} %d\n", bound, m.buckets[i])
	}
	fmt.Fprintf(c.Writer, "http_request_duration_seconds_bucket{le=\"+Inf\"} %d\nhttp_request_duration_seconds_sum %g\nhttp_request_duration_seconds_count %d\n", m.count, m.seconds, m.count)
}
