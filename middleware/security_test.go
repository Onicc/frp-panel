package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", LoginRateLimit(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	for attempt := 1; attempt <= 6; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/login", nil)
		request.RemoteAddr = "192.0.2.10:1234"
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if attempt <= 5 && response.Code != http.StatusNoContent {
			t.Fatalf("attempt %d unexpectedly rejected", attempt)
		}
		if attempt == 6 && response.Code != http.StatusTooManyRequests {
			t.Fatalf("sixth attempt returned %d", response.Code)
		}
	}
}
