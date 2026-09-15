package api

import (
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestServiceStopsItsListener(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	service := NewApiService(listener, router, true)
	done := make(chan struct{})
	go func() {
		service.Run()
		close(done)
	}()

	response, err := http.Get("http://" + listener.Addr().String() + "/health") // #nosec G107 -- test-only loopback listener
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	service.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("API service did not stop")
	}
}
