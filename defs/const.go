package defs

import (
	"time"
)

const (
	AuthorizationKey = "authorization"
	MsgKey           = "msg"
	UserIDKey        = "x-frp-panel-userid"
	EndpointKey      = "endpoint"
	IpAddrKey        = "ipaddr"
	HeaderKey        = "header"
	MethodKey        = "method"
	UAKey            = "User-Agent"
	ContentTypeKey   = "Content-Type"
	TraceIDKey       = "TraceID"
	TokenKey         = "token"
	FRPAuthTokenKey  = "token"
	ErrKey           = "err"
	UserInfoKey      = "x-frp-panel-userinfo"
	FRPClientIDKey   = "x-frp-panel-client-id"
)

const (
	ErrInvalidRequest   = "invalid request"
	ErrUserInfoNotValid = "user info not valid"
	ErrInternalError    = "internal error"
	ErrParamNotValid    = "param not valid"
	ErrDB               = "database error"
	ErrNotFound         = "data not found"
	ErrCodeNotFound     = "code not found"
	ErrCodeAlreadyUsed  = "code already used"
)

const (
	ReqSuccess = "success"
)

const (
	TimeLayout = time.RFC3339
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFailed  = "failed"
	StatusDone    = "done"
)

const (
	CliTypeServer = "server"
	CliTypeClient = "client"
)

type AppRole string

const (
	AppRole_Client AppRole = CliTypeClient
	AppRole_Server AppRole = CliTypeServer
	AppRole_Master AppRole = "master"
)

const (
	LocalHost            = "127.0.0.1"
	FRP_Plugin_Multiuser = "multiuser"
)

const (
	PullConfigDuration           = 30 * time.Second
	PushProxyInfoDuration        = 30 * time.Second
	PullClientWorkersDuration    = 30 * time.Second
	PullClientWireGuardsDuration = 30 * time.Second

	ReportWireGuardRuntimeInfoDuration = 60 * time.Second

	AppStartTimeout = 5 * time.Minute
)

const (
	CurEnvPath = ".env"
	SysEnvPath = "/etc/frp-panel/master.env"
)

const (
	DBRoleDefault = "default"
	DBRoleRam     = "ram"
)

const (
	DBTypeSQLite3  = "sqlite3"
	DBTypePostgres = "postgres"
)

const (
	UserRole_Owner    = "owner"
	UserRole_Admin    = "admin"
	UserRole_Operator = "operator"
	UserRole_Viewer   = "viewer"
	UserRole_Normal   = UserRole_Viewer
	CapFileName       = "workerd.capnp"
	WorkerInfoPath    = "workers"
	WorkerCodePath    = "src"
	DBTypeSqlite      = "sqlite"

	DefaultHostName       = "127.0.0.1"
	DefaultNodeName       = "default"
	DefaultExternalPath   = "/"
	DefaultEntry          = "entry.js"
	DefaultSocketTemplate = "unix-abstract:/tmp/frpp-worker-%s.sock"
	DefaultCode           = `export default {
  async fetch(req, env) {
    try {
		let resp = new Response("worker: " + req.url + " is online! -- " + new Date())
		return resp
	} catch(e) {
		return new Response(e.stack, { status: 500 })
	}
  }
};`

	DefaultConfigTemplate = `using Workerd = import "/workerd/workerd.capnp";

const config :Workerd.Config = (
  services = [
    (name = "{{.WorkerId}}", worker = .v{{.WorkerId}}Worker),
  ],

  sockets = [
    (
      name = "{{.WorkerId}}",
      address = "{{.Socket.Address}}",
      http=(),
      service="{{.WorkerId}}"
    ),
  ]
);

const v{{.WorkerId}}Worker :Workerd.Worker = (
  modules = [
    (name = "{{.CodeEntry}}", esModule = embed "src/{{.CodeEntry}}"),
  ],
  compatibilityDate = "2023-04-03",
);`
)

type TokenStatus string

const (
	TokenStatusActive   TokenStatus = "active"
	TokenStatusInactive TokenStatus = "inactive"
	TokenStatusRevoked  TokenStatus = "revoked"
)

const (
	KeyNodeName    = "node_name"
	KeyNodeSecret  = "node_secret"
	KeyNodeProto   = "node_proto"
	KeyWorkerProto = "worker_proto"
)

type WorkerStatus string

const (
	WorkerStatus_Unknown  WorkerStatus = "unknown"
	WorkerStatus_Running  WorkerStatus = "running"
	WorkerStatus_Inactive WorkerStatus = "inactive"
)

const (
	FrpProxyAnnotationsKey_Ingress           = "ingress"
	FrpProxyAnnotationsKey_WorkerId          = "worker_id"
	FrpProxyAnnotationsKey_LoadBalancerGroup = "load_balancer_group"
)

const (
	PlaceholderPrivateKey         = "<REPLACE_WITH_YOUR_PRIVATE_KEY>"
	PlaceholderPeerVPNAddressCIDR = "<PEER_VPN_IP_ADDRESS/PREFIX>"
)

const (
	DefaultDeviceMTU           uint32 = 1400
	DefaultPersistentKeepalive uint32 = 25
)

var VaalaMagicBytes = []byte("vaala-ping")

const VaalaMagicBytesCookie = uint32(1630367849)

const (
	DefaultWSHandlerPath = "/api/frp-panel-transport/ws"
)
