package rpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/Onicc/frp-panel/utils/wsgrpc"
	"github.com/imroc/req/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

type masterClient struct {
	cli         pb.MasterClient
	inited      bool
	appInstance app.Application
}

func (m *masterClient) Call() pb.MasterClient {
	if !m.inited {
		m.cli = newMasterCli(m.appInstance)
		m.inited = true
	}
	return m.cli
}

func NewMasterCli(appInstance app.Application) *masterClient {
	logger.Logger(context.Background()).Debugf("creating new master client")
	return &masterClient{
		inited:      false,
		appInstance: appInstance,
	}
}

func newMasterCli(appInstance app.Application) pb.MasterClient {
	connInfo := conf.GetRPCConnInfo(appInstance.GetConfig())
	ctx := context.Background()

	opt := []grpc.DialOption{}

	switch connInfo.Scheme {
	case conf.GRPC:
		if appInstance.GetConfig().Client.TLSRpc {
			logger.Logger(ctx).Infof("use tls rpc")
			opt = append(opt, grpc.WithTransportCredentials(appInstance.GetRPCCred()))
		} else {
			logger.Logger(ctx).Infof("use insecure rpc")
			opt = append(opt, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	case conf.WS, conf.WSS:
		logger.Logger(ctx).Infof("use ws/wss rpc")

		wsURL := fmt.Sprintf("%s://%s/wsgrpc", connInfo.Scheme, connInfo.Host)
		header := http.Header{}
		wsDialer := wsgrpc.WebsocketDialer(wsURL,
			header,
			appInstance.GetConfig().Client.TLSInsecureSkipVerify,
			logger.Logger(ctx),
		)
		opt = append(opt, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(wsDialer))
	}

	logger.Logger(ctx).Debugf("creating new grpc client to [%s]", utils.MarshalForJson(connInfo))
	conn, err := grpc.NewClient(connInfo.Host, opt...)

	if err != nil {
		logger.Logger(ctx).Fatalf("did not connect: %v", err)
	}

	logger.Logger(ctx).Debugf("grpc client created")

	return pb.NewMasterClient(conn)
}

func httpCli(insecureSkipVerify bool) *req.Client {
	c := req.C()
	c.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: insecureSkipVerify} // #nosec G402 -- explicit opt-in compatibility flag
	return c
}

func GetClientCert(appInstance app.Application, clientID, clientSecret string, clientType pb.ClientType) []byte {
	apiEndpoint := conf.GetAPIURL(appInstance.GetConfig())
	c := httpCli(appInstance.GetConfig().Client.TLSInsecureSkipVerify)

	rawReq, err := proto.Marshal(&pb.GetClientCertRequest{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		ClientType:   clientType,
	})
	if err != nil {
		return nil
	}
	r, err := c.R().SetHeader("Content-Type", "application/x-protobuf").
		SetBodyBytes(rawReq).Post(apiEndpoint + "/api/v1/auth/cert")
	if err != nil {
		return nil
	}

	resp := &pb.GetClientCertResponse{}
	err = proto.Unmarshal(r.Bytes(), resp)
	if err != nil {
		return nil
	}
	return resp.Cert
}
