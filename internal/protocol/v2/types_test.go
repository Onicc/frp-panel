package v2

import "testing"

func TestDesiredStateRevision(t *testing.T) {
	state := DesiredState{Revision: 5}
	if state.ShouldApply(5) || state.ShouldApply(6) || !state.ShouldApply(4) {
		t.Fatal("desired state revision ordering is incorrect")
	}
}

func TestHelloRequiresVersionAndIdentity(t *testing.T) {
	if err := (Hello{Protocol: Version, NodeID: "node-1", AgentVersion: "v2"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (Hello{Protocol: "1"}).Validate(); err == nil {
		t.Fatal("invalid hello accepted")
	}
}
