// Package observability provides shared, bounded-cardinality API instrumentation.
package observability

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	service  string
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	handler  http.Handler
}

func New(service string) *Metrics {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "API requests by service, method, route, and HTTP status."}, []string{"service", "method", "route", "status"})
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "API response latency in seconds.", Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30}}, []string{"service", "method", "route"})
	registry.MustRegister(requests, duration)
	return &Metrics{service: service, requests: requests, duration: duration, handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{})}
}
func (m *Metrics) Handler(c *gin.Context) { m.handler.ServeHTTP(c.Writer, c.Request) }
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		path := c.Request.URL.Path
		if path == "/metrics" || path == "/healthz" || path == "/readyz" || path == "/" || strings.HasPrefix(path, "/swagger/") {
			return
		}
		route := c.FullPath()
		if m.service == "api-gateway" {
			route = GatewayRoute(path)
		}
		if route == "" {
			route = "unmatched"
		}
		method := c.Request.Method
		switch method {
		case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		default:
			method = "OTHER"
		}
		elapsed := time.Since(started).Seconds()
		m.requests.WithLabelValues(m.service, method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.duration.WithLabelValues(m.service, method, route).Observe(elapsed)
		record, _ := json.Marshal(map[string]any{"event": "http_request", "service": m.service, "method": method, "route": route, "status": c.Writer.Status(), "duration_ms": elapsed * 1000})
		log.Print(string(record))
	}
}

// Gateway wildcards are mapped to known service route templates. Never label raw paths.
func GatewayRoute(path string) string {
	if !strings.HasPrefix(path, "/api/") {
		return "unmatched"
	}
	path = strings.TrimPrefix(path, "/api")
	switch path {
	case "/user/signup", "/user/login", "/user/logout", "/user/refresh", "/feed/create", "/feed/", "/feed/paginated", "/comment/create", "/comment/paginated":
		return path
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 2 && parts[0] == "feed" {
		if _, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
			return "/feed/:id"
		}
	}
	return "unmatched"
}
