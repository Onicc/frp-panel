package clientrpc

import (
	"context"
	"io"
	"time"

	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/google/uuid"
)

func registerWithMaster(ctx context.Context, recvStream pb.Master_ServerSendClient, event pb.Event, componentID, componentSecret string) {
	logger.Logger(ctx).Info("register managed component with Master")
	for {
		err := recvStream.Send(&pb.ClientMessage{
			Event:     event,
			ClientId:  componentID,
			SessionId: uuid.New().String(),
			Secret:    componentSecret,
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Logger(ctx).WithError(err).Warnf("cannot send, sleep 3s and retry")
			if !waitForRetry(ctx) {
				return
			}
			continue
		}

		resp, err := recvStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			if ctx.Err() == nil {
				logger.Logger(ctx).WithError(err).Warn("cannot receive registration response")
			}
			return
		}

		if resp.GetEvent() == event {
			logger.Logger(ctx).Infof("managed component registered with Master, component ID: [%s]", resp.GetClientId())
			break
		}
	}
}

func runClientRPCHandler(ctx context.Context, appInstance app.Application, recvStream pb.Master_ServerSendClient, clientID string,
	clientHandleServerSend func(appInstance app.Application, req *pb.ServerMessage) *pb.ClientMessage) {
	for {
		select {
		case <-ctx.Done():
			logger.Logger(ctx).Infof("finish rpc client")
			_ = recvStream.CloseSend()
			return
		default:
			resp, err := recvStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				if ctx.Err() == nil {
					logger.Logger(ctx).WithError(err).Errorf("cannot receive, retrying")
					_ = waitForRetry(ctx)
				}
				return
			}
			if resp == nil {
				continue
			}
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Logger(ctx).Errorf("catch panic, err: %v", err)
					}
				}()
				msg := clientHandleServerSend(appInstance, resp)
				if msg == nil {
					return
				}
				msg.ClientId = clientID
				msg.SessionId = resp.SessionId
				recvStream.Send(msg)
				logger.Logger(ctx).Infof("client resp received: %s", resp.GetClientId())
			}()
		}
	}
}

func startClientRpcHandler(ctx context.Context, appInstance app.Application, client app.MasterClient, clientID, clientSecret string, event pb.Event,
	clientHandleServerSend func(appInstance app.Application, req *pb.ServerMessage) *pb.ClientMessage) {
	logger.Logger(ctx).Infof("start to run rpc client")
	for {
		select {
		case <-ctx.Done():
			logger.Logger(ctx).Infof("finish rpc client")
			return
		default:
			recvStream, err := client.Call().ServerSend(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Logger(ctx).WithError(err).Errorf("cannot recv, sleep 3s and retry")
				if !waitForRetry(ctx) {
					return
				}
				continue
			}

			registerWithMaster(ctx, recvStream, event, clientID, clientSecret)
			runClientRPCHandler(ctx, appInstance, recvStream, clientID, clientHandleServerSend)
		}
	}
}

func waitForRetry(ctx context.Context) bool {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
