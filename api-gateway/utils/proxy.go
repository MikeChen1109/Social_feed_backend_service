package utils

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
	"time"
)

var proxyClient = &http.Client{Timeout: 10 * time.Second}

func ProxyRequest(c *gin.Context, targetBaseURL string) {
	proxyRequest(c, targetBaseURL, proxyClient)
}

func proxyRequest(c *gin.Context, targetBaseURL string, client *http.Client) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api")
	targetURL := targetBaseURL + path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request creation failed"})
		return
	}
	req.Header = c.Request.Header.Clone()
	// Hop-by-hop headers must not be forwarded to another HTTP connection.
	for _, name := range strings.Split(req.Header.Get("Connection"), ",") {
		req.Header.Del(strings.TrimSpace(name))
	}
	for _, name := range []string{"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "TE", "Trailer", "Transfer-Encoding", "Upgrade"} {
		req.Header.Del(name)
	}
	resp, err := client.Do(req)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
		}
		c.JSON(status, gin.H{"error": "Upstream request failed"})
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Upstream response incomplete"})
		return
	}
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
