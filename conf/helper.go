package conf

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"time"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

func RPCListenAddr(cfg Config) string {
	return fmt.Sprintf(":%d", cfg.Master.RPCPort)
}

func rpcCallAddr(cfg Config) string {
	return fmt.Sprintf("%s:%d", cfg.Master.RPCHost, cfg.Master.RPCPort)
}

func JWTSecret(cfg Config) string {
	hash := sha256.Sum256([]byte("frp-panel:v2:" + cfg.App.GlobalSecret))
	return fmt.Sprintf("%x", hash[:])
}

func MasterAPIListenAddr(cfg Config) string {
	return fmt.Sprintf(":%d", cfg.Master.APIPort)
}

func ServerAPIListenAddr(cfg Config) string {
	return fmt.Sprintf("%s:%d", defs.LocalHost, cfg.Server.APIPort)
}

func FRPsAuthOption(cfg Config) v1.HTTPPluginOptions {
	authUrl := fmt.Sprintf("http://%s:%d/auth", defs.LocalHost, cfg.Server.APIPort)
	parsedUrl, err := url.Parse(authUrl)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Fatalf("parse auth url error")
	}

	return v1.HTTPPluginOptions{
		Name: defs.FRP_Plugin_Multiuser,
		Ops:  []string{"Login"},
		Addr: parsedUrl.Host,
		Path: parsedUrl.Path,
	}
}

func GetJWTWithPayload(cfg Config, uid int, payload map[string]interface{}) (string, error) {
	payload[defs.UserIDKey] = uid
	return utils.GetJwtTokenFromMap(JWTSecret(cfg),
		time.Now().Unix(),
		int64(cfg.App.CookieAge),
		payload)
}

func PermissionsForRole(role string) []defs.APIPermission {
	permissions := []defs.APIPermission{
		{Method: "GET", Path: "*"},
		{Method: "POST", Path: `^/api/v1/(user/(get|update)|platform/clientsstatus|client/(get|list)|server/(get|list)|proxy/(get_by_cid|get_by_sid|list_configs|get_config)|worker/(get|status|list|get_ingress)|wg/.*(get|list|topology))$`},
	}
	switch role {
	case defs.UserRole_Owner, defs.UserRole_Admin:
		return []defs.APIPermission{{Method: "*", Path: "*"}}
	case defs.UserRole_Operator:
		permissions = append(permissions,
			defs.APIPermission{Method: "POST", Path: `^/api/v1/(client|server|frpc|frps|proxy|wg|worker)(/.*)?$`},
			defs.APIPermission{Method: "POST", Path: `^/api/v2/(enrollments|server-enrollments|node-routes)$`},
			defs.APIPermission{Method: "DELETE", Path: `^/api/v2/node-routes$`},
		)
	}
	return permissions
}

func GetCommonJWT(cfg Config, uid int) string {
	token, _ := GetJWTWithPayload(cfg, uid, map[string]interface{}{})
	return token
}

func GetCommonJWTWithExpireTime(cfg Config, uid int64, expSec int) string {
	token, _ := utils.GetJwtTokenFromMap(JWTSecret(cfg),
		time.Now().Unix(),
		int64(expSec),
		map[string]interface{}{defs.UserIDKey: uid})
	return token
}

func GetAPIURL(cfg Config) string {

	if len(cfg.Client.APIUrl) != 0 {
		return cfg.Client.APIUrl
	}

	return fmt.Sprintf("%s://%s:%d", cfg.Master.APIScheme, cfg.Master.APIHost, cfg.Master.APIPort)
}

func GetCertTemplate(cfg Config) *x509.Certificate {
	now := time.Now()
	return &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName:         cfg.Master.APIHost,
			Country:            []string{"CN"},
			Organization:       []string{"frp-panel"},
			OrganizationalUnit: []string{"frp-panel"},
		},
		SignatureAlgorithm:    x509.SHA512WithRSA,
		DNSNames:              []string{cfg.Master.APIHost},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		NotBefore:             now,
		NotAfter:              now.AddDate(10, 0, 0),
		SubjectKeyId:          []byte{102, 114, 112, 45, 112, 97, 110, 101, 108},
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyAgreement |
			x509.KeyUsageDataEncipherment,
	}
}

type LisOpt struct {
	MuxLis net.Listener
	ApiLis net.Listener
	RunAPI bool
}

func GetListener(c context.Context, cfg Config) LisOpt {
	runAPI := RPCListenAddr(cfg) != MasterAPIListenAddr(cfg)

	muxLis, err := net.Listen("tcp", RPCListenAddr(cfg))
	if err != nil {
		logger.Logger(c).WithError(err).Fatalf("failed to listen: %v", RPCListenAddr(cfg))
	}

	opt := LisOpt{
		MuxLis: muxLis,
		RunAPI: runAPI,
	}

	if runAPI {
		apiLis, err := net.Listen("tcp", MasterAPIListenAddr(cfg))
		if err != nil {
			logger.Logger(c).WithError(err).Warnf("failed to listen: %v, but mux server can handle http api", MasterAPIListenAddr(cfg))
		}
		opt.ApiLis = apiLis
	}

	if !runAPI {
		opt.ApiLis = nil
	}

	return opt
}

type ConnInfo struct {
	Host   string
	Scheme Scheme
}

type Scheme string

const (
	GRPC Scheme = "grpc"
	WS   Scheme = "ws"
	WSS  Scheme = "wss"
)

func GetRPCConnInfo(cfg Config) ConnInfo {
	rpcUrl := cfg.Client.RPCUrl

	if len(rpcUrl) == 0 {
		return ConnInfo{
			Host:   rpcCallAddr(cfg),
			Scheme: GRPC,
		}
	}

	parsedUrl, err := url.Parse(rpcUrl)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Fatalf("parse rpc url error")
	}

	return ConnInfo{
		Host:   parsedUrl.Host,
		Scheme: Scheme(parsedUrl.Scheme),
	}
}
