package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"app/internal/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestTimeout_AllowsFastRequests(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(middleware.Timeout(100 * time.Millisecond))
	router.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestTimeout_ReturnsGatewayTimeoutWhenExceeded(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(middleware.Timeout(50 * time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
			return
		case <-time.After(200 * time.Millisecond):
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusGatewayTimeout, rec.Code, rec.Body.String())
	}
}

func TestTimeout_PropagatesContextToHandler(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.Use(middleware.Timeout(100 * time.Millisecond))
	router.GET("/ctx", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		if !ok {
			t.Error("expected request context to have a deadline")
			return
		}
		if time.Until(deadline) <= 0 {
			t.Error("expected deadline in the future")
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ctx", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
