package monitoring

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestMetricsCountConcurrentRequestsAndExcludeProbes(t *testing.T) {
	m := New()
	r := gin.New()
	r.Use(m.Middleware(), gin.Recovery())
	r.GET("/metrics", m.Handler)
	r.GET("/panic", func(c *gin.Context) { panic("test") })
	r.GET("/healthz", func(c *gin.Context) { c.Status(200) })
	r.GET("/work/:id", func(c *gin.Context) { c.Status(403) })
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/work/123", nil))
		}()
	}
	wg.Wait()
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/panic", nil))
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/healthz", nil))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	for _, expected := range []string{"http_requests_total 21\n", "http_errors_total 21\n", "http_request_duration_seconds_count 21\n", "http_request_duration_seconds_bucket{le=\"+Inf\"} 21\n"} {
		if !strings.Contains(w.Body.String(), expected) {
			t.Fatalf("missing %q in %s", expected, w.Body.String())
		}
	}
}
