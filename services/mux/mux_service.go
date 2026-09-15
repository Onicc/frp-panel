package mux

import (
	"context"
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type MuxServer interface {
	Run()
	Stop()
}

type muxImpl struct {
	srv *http.Server
	lis net.Listener
	tls bool
}

func NewMux(grpcServer, apiServer http.Handler, lis net.Listener, creds *tls.Config) MuxServer {
	tlsServer := grpcHandlerFunc(grpcServer, apiServer)
	tlsServer.TLSConfig = creds
	return &muxImpl{
		srv: tlsServer,
		lis: lis,
		tls: creds != nil,
	}
}

func (m *muxImpl) Run() {
	if m.tls {
		if err := m.srv.ServeTLS(m.lis, "", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("TLS mux server stopped: %v", err)
		}
	} else {
		if err := m.srv.Serve(m.lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP mux server stopped: %v", err)
		}
	}
}

func (m *muxImpl) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := m.srv.Shutdown(ctx); err != nil {
		_ = m.srv.Close()
	}
}

func grpcHandlerFunc(grpcServer http.Handler, httpHandler http.Handler) *http.Server {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(true)
	protocols.SetUnencryptedHTTP2(true)
	return &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("proto major: %d,  %s , %s\n", r.ProtoMajor, r.RequestURI, r.Header.Get("Content-Type"))
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	}), Protocols: protocols, ReadHeaderTimeout: 30 * time.Second}
}
