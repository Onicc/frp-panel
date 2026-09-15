package defs

import "github.com/Onicc/frp-panel/pb"

type Connector struct {
	CliID   string
	Conn    pb.Master_ServerSendServer
	CliType string
}
