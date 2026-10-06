package observability

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRoutesStatusesLatencyAndProbeExclusion(t *testing.T) {
	m := New("feed-service")
	r := gin.New()
	r.Use(m.Middleware(), gin.Recovery())
	r.GET("/metrics", m.Handler)
	r.GET("/feed/:id", func(c *gin.Context) { c.Status(403) })
	r.GET("/healthz", func(c *gin.Context) { c.Status(200) })
	r.GET("/panic", func(c *gin.Context) { panic("test") })
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/feed/123?token=secret", nil))
		}()
	}
	wg.Wait()
	for _, path := range []string{"/healthz", "/swagger/index.html", "/panic", "/unknown-user-data"} {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	text := w.Body.String()
	for _, expected := range []string{`http_requests_total{method="GET",route="/feed/:id",service="feed-service",status="403"} 20`, `http_requests_total{method="GET",route="/panic",service="feed-service",status="500"} 1`, `http_request_duration_seconds_count{method="GET",route="/feed/:id",service="feed-service"} 20`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %s in %s", expected, text)
		}
	}
	for _, sensitive := range []string{"/feed/123", "secret", "unknown-user-data", "/healthz", "/swagger"} {
		if strings.Contains(text, sensitive) {
			t.Fatalf("unexpected raw data/probe: %s", sensitive)
		}
	}
}
func TestGatewayNormalizesRoutes(t *testing.T) {
	for path, want := range map[string]string{"/api/feed/123": "/feed/:id", "/api/feed/456": "/feed/:id", "/api/user/login": "/user/login", "/api/feed/paginated": "/feed/paginated", "/api/feed/raw-user-text": "unmatched", "/anything": "unmatched"} {
		if got := GatewayRoute(path); got != want {
			t.Fatalf("%s: %s != %s", path, got, want)
		}
	}
}
