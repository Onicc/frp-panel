package client

import (
	"testing"

	"github.com/Onicc/frp-panel/pb"
)

func TestPrivilegedEventClassification(t *testing.T) {
	if !isFunctionEvent(pb.Event_EVENT_CREATE_WORKER) || isFunctionEvent(pb.Event_EVENT_UPDATE_FRPC) {
		t.Fatal("function event classification is incorrect")
	}
	if !isWireGuardEvent(pb.Event_EVENT_CREATE_WIREGUARD) || isWireGuardEvent(pb.Event_EVENT_START_PTY_CONNECT) {
		t.Fatal("WireGuard event classification is incorrect")
	}
}
