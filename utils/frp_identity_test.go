package utils

import "testing"

func TestFRPClientUserSeparatesPhysicalClients(t *testing.T) {
	first := FRPClientUser("owner", "owner.c.one")
	second := FRPClientUser("owner", "owner.c.two")
	if first == second || first == "owner" || second == "owner" {
		t.Fatalf("FRP Client namespaces must be unique: %q %q", first, second)
	}
	if first != FRPClientUser("owner", "owner.c.one") {
		t.Fatal("FRP Client namespace is not deterministic")
	}
	if FRPAccountName(first) != "owner" || FRPAccountName("owner") != "owner" {
		t.Fatal("FRPS authentication did not retain the account name")
	}
	if first+".ssh" == second+".ssh" {
		t.Fatal("same Tunnel name on different Clients produced one wire name")
	}
}

func TestManagedFRPCRetriesUntilServerIsReady(t *testing.T) {
	cfg := NewBaseFRPClientUserAuthConfig("edge.example.test", 7001, "owner.client", "token")
	if cfg.LoginFailExit == nil || *cfg.LoginFailExit {
		t.Fatal("managed FRPC must retry a Server that starts after the Client")
	}
}
