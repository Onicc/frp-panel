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

// func clientHandleServerSend(req *pb.ServerMessage) *pb.ClientMessage {
// 	logger.Logger(c).Infof("client get a server message, origin is: [%+v]", req)
// 	return &pb.ClientMessage{
// 		Event:     pb.Event_EVENT_DATA,
// 		ClientId:  req.ClientId,
// 		SessionId: req.SessionId,
// 		Data:      req.Data,
// 	}
// }

func registClientToMaster(ctx context.Context, appInstance app.Application, recvStream pb.Master_ServerSendClient, event pb.Event, clientID, clientSecret string) {
	logger.Logger(ctx).Infof("start to regist client to master")
	for {
		err := recvStream.Send(&pb.ClientMessage{
			Event:     event,
			ClientId:  clientID,
			SessionId: uuid.New().String(),
			Secret:    clientSecret,
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
			logger.Logger(ctx).Infof("client get server register envent success, clientID: %s", resp.GetClientId())
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

			registClientToMaster(ctx, appInstance, recvStream, event, clientID, clientSecret)
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
