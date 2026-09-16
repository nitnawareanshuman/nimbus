package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestShortenRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		body          string
		expectedError string
	}{
		{
			name:          "invalid JSON",
			body:          `{`,
			expectedError: "invalid JSON body",
		},
		{
			name:          "missing URL",
			body:          `{}`,
			expectedError: "url is required",
		},
		{
			name:          "blank URL",
			body:          `{"url":"   "}`,
			expectedError: "url is required",
		},
		{
			name:          "invalid URL",
			body:          `{"url":"not a url"}`,
			expectedError: "invalid URL",
		},
		{
			name:          "unsupported scheme",
			body:          `{"url":"ftp://example.com/file"}`,
			expectedError: "URL must use http or https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(
				http.MethodPost,
				"/shorten",
				strings.NewReader(tt.body),
			)
			ctx.Request.Header.Set("Content-Type", "application/json")

			h.Shorten(ctx)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}

			if !strings.Contains(recorder.Body.String(), tt.expectedError) {
				t.Fatalf(
					"expected response to contain %q, got %s",
					tt.expectedError,
					recorder.Body.String(),
				)
			}
		})
	}
}
