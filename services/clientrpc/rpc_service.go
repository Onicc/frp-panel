package clientrpc

import (
	"context"
	"sync"

	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
)

type ClientRPCHandler interface {
	Run()
	Stop()
	GetCli() pb.MasterClient
}

type clientRPCHandler struct {
	appInstance  app.Application
	rpcClient    app.MasterClient
	handerFunc   func(appInstance app.Application, req *pb.ServerMessage) *pb.ClientMessage
	clientID     string
	clientSecret string
	event        pb.Event
	ctx          context.Context
	cancel       context.CancelFunc
	stopOnce     sync.Once
}

func NewClientRPCHandler(
	appInstance app.Application,
	clientID,
	clientSecret string,
	event pb.Event,
	handerFunc func(appInstance app.Application, req *pb.ServerMessage) *pb.ClientMessage,
) app.ClientRPCHandler {
	rpcCli := appInstance.GetMasterCli()
	ctx, cancel := context.WithCancel(context.Background())
	return &clientRPCHandler{
		appInstance:  appInstance,
		rpcClient:    rpcCli,
		handerFunc:   handerFunc,
		clientID:     clientID,
		clientSecret: clientSecret,
		event:        event,
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (s *clientRPCHandler) Run() {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger(context.Background()).Fatalf("client rpc handler panic: %v", err)
		}
	}()

	startClientRpcHandler(s.ctx, s.appInstance, s.rpcClient, s.clientID, s.clientSecret, s.event, s.handerFunc)
}

func (s *clientRPCHandler) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
	})
}

func (s *clientRPCHandler) GetCli() app.MasterClient {
	return s.rpcClient
}
