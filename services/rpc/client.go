package rpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func CallClientWrapper[R common.RespType](c *app.Context, clientID string, event pb.Event, req proto.Message, resp *R) error {
	cresp, err := CallClient(c, clientID, event, req)
	if err != nil {
		return err
	}

	protoMsgRef, ok := any(resp).(protoreflect.ProtoMessage)
	if !ok {
		return fmt.Errorf("type does not implement protoreflect.ProtoMessage")
	}

	return proto.Unmarshal(cresp.GetData(), protoMsgRef)
}

func CallClient(ctx *app.Context, clientID string, event pb.Event, msg proto.Message) (*pb.ClientMessage, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return CallClientWithContext(ctx, requestCtx, clientID, event, msg)
}

// CallClientWithContext is the bounded variant used by status/reconciliation
// paths. The legacy CallClient entry point keeps the existing application
// context semantics while still inheriting a finite timeout.
func CallClientWithContext(appCtx *app.Context, requestCtx context.Context, clientID string, event pb.Event, msg proto.Message) (*pb.ClientMessage, error) {
	manager := appCtx.GetApp().GetClientsManager()
	if manager == nil {
		return nil, fmt.Errorf("client manager is unavailable")
	}
	sender := manager.Get(clientID)
	if sender == nil {
		logger.Logger(appCtx).Errorf("cannot get client, id: [%s]", clientID)
		return nil, fmt.Errorf("cannot get client, id: [%s]", clientID)
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot marshal")
		return nil, err
	}

	req := &pb.ServerMessage{
		Event:     event,
		Data:      data,
		SessionId: uuid.New().String(),
		ClientId:  clientID,
	}

	responseChannel := make(chan *pb.ClientMessage, 1)
	appInstance := appCtx.GetApp()
	appInstance.GetClientRecvMap().Store(req.SessionId, responseChannel)
	defer appInstance.GetClientRecvMap().Delete(req.SessionId)
	err = sender.Conn.Send(req)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot send")
		manager.Remove(clientID)
		return nil, err
	}
	respChAny, ok := appInstance.GetClientRecvMap().Load(req.SessionId)
	if !ok {
		logger.Logger(appCtx).Errorf("cannot load response channel")
		return nil, fmt.Errorf("response channel unavailable")
	}

	respCh, ok := respChAny.(chan *pb.ClientMessage)
	if !ok {
		logger.Logger(appCtx).Errorf("cannot cast response channel")
		return nil, fmt.Errorf("response channel has invalid type")
	}

	if requestCtx == nil {
		requestCtx = context.Background()
	}
	select {
	case resp := <-respCh:
		if resp.Event == pb.Event_EVENT_ERROR {
			return nil, fmt.Errorf("client return error: %s", resp.Data)
		}
		return resp, nil
	case <-requestCtx.Done():
		return nil, requestCtx.Err()
	}
}

func Recv(appInstance app.Application, clientID string) chan bool {
	done := make(chan bool, 1)
	go func() {
		c := context.Background()
		log := logger.Logger(c).WithField("clientID", clientID)
		manager := appInstance.GetClientsManager()
		if manager == nil {
			done <- true
			return
		}
		for {
			reciver := manager.Get(clientID)
			if reciver == nil {
				log.Errorf("cannot get client; ending receive loop")
				done <- true
				return
			}
			resp, err := reciver.Conn.Recv()
			if err == io.EOF {
				log.Infof("finish client recv")
				done <- true
				return
			}
			if err != nil {
				log.WithError(err).Errorf("cannot recv, usually means client disconnect")
				done <- true
				return
			}

			respChAny, ok := appInstance.GetClientRecvMap().Load(resp.SessionId)
			if !ok {
				log.Errorf("cannot load")
				continue
			}

			respCh, ok := respChAny.(chan *pb.ClientMessage)
			if !ok {
				log.Errorf("cannot cast")
				continue
			}
			log.Debugf("recv success, resp: %+v", resp)
			respCh <- resp
		}
	}()
	return done
}
