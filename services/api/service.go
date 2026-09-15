package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ApiService interface {
	Run()
	Stop()
}

type server struct {
	srv    *http.Server
	addr   net.Listener
	enable bool
}

var (
	_ ApiService = (*server)(nil)
)

func NewApiService(listen net.Listener, router *gin.Engine, enable bool) *server {
	return &server{
		srv:    &http.Server{Handler: router, ReadHeaderTimeout: 30 * time.Second},
		addr:   listen,
		enable: enable,
	}
}

func (s *server) Run() {
	// 如果完全使用mux，可以不启动
	if !s.enable {
		return
	}
	if err := s.srv.Serve(s.addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return
	}
}

func (s *server) Stop() {
	if !s.enable {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		_ = s.srv.Close()
	}
}
