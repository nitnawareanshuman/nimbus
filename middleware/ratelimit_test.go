package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func resetVisitors() {
	mu.Lock()
	defer mu.Unlock()

	visitors = make(map[string]*visitor)
}

func TestRateLimitAllowsRequestsWithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	router := gin.New()
	router.Use(RateLimit())

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < requestsPerMinute; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		req.RemoteAddr = "192.0.2.1:1234"

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"request %d: expected status %d, got %d",
				i+1,
				http.StatusOK,
				recorder.Code,
			)
		}
	}
}

func TestRateLimitRejectsRequestsOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	router := gin.New()
	router.Use(RateLimit())

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < requestsPerMinute; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		req.RemoteAddr = "192.0.2.2:1234"

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf(
				"request %d unexpectedly failed with %d",
				i+1,
				recorder.Code,
			)
		}
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req.RemoteAddr = "192.0.2.2:1234"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusTooManyRequests,
			recorder.Code,
		)
	}

	expected := "too many requests"

	if body := recorder.Body.String(); !contains(body, expected) {
		t.Fatalf(
			"expected response to contain %q, got %s",
			expected,
			body,
		)
	}
}

func TestRateLimitSeparatesVisitors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	resetVisitors()

	router := gin.New()
	router.Use(RateLimit())

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < requestsPerMinute; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/test",
			nil,
		)

		req.RemoteAddr = "192.0.2.3:1234"

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/test",
		nil,
	)

	req.RemoteAddr = "192.0.2.4:1234"

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected different visitor to be allowed, got %d: %s",
			recorder.Code,
			fmt.Sprint(recorder.Body.String()),
		)
	}
}

func contains(value, target string) bool {
	for i := 0; i+len(target) <= len(value); i++ {
		if value[i:i+len(target)] == target {
			return true
		}
	}

	return false
}