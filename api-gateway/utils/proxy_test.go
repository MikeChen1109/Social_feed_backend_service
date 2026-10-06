package utils

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProxyTimeoutCancelsUpstream(t *testing.T) {
	canceled := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done(); close(canceled) }))
	defer upstream.Close()
	r := gin.New()
	r.GET("/api/test", func(c *gin.Context) { proxyRequest(c, upstream.URL, &http.Client{Timeout: 100 * time.Millisecond}) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/test", nil))
	if w.Code != 504 {
		t.Fatalf("expected 504, got %d", w.Code)
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("upstream not canceled")
	}
}
func TestProxyPropagatesRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := gin.New()
	r.GET("/api/test", func(c *gin.Context) { proxyRequest(c, "http://localhost:1", &http.Client{}) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/test", nil).WithContext(ctx))
	if w.Code != 502 {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}
