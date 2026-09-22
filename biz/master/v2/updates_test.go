package v2

import (
	"strings"
	"testing"
	"time"

	"github.com/Onicc/frp-panel/internal/containerupdate"
	"github.com/Onicc/frp-panel/models"
)

func TestMasterRestartWaitsForLauncherDecision(t *testing.T) {
	t.Setenv("FRP_PANEL_UPDATE_DIR", t.TempDir())
	opID := "pending-master"
	commit := strings.Repeat("a", 40)
	if err := containerupdate.WriteStatus(containerupdate.Status{OperationID: opID, Commit: commit, State: "staged"}); err != nil {
		t.Fatal(err)
	}
	if err := containerupdate.WriteManifest(containerupdate.PendingPath(), containerupdate.Manifest{OperationID: opID, Commit: commit}); err != nil {
		t.Fatal(err)
	}
	state, _ := masterRestartDecision(opID)
	if state != "" {
		t.Fatalf("premature terminal state before launcher readiness: %s", state)
	}
	if err := containerupdate.WriteStatus(containerupdate.Status{OperationID: opID, Commit: commit, State: "succeeded"}); err != nil {
		t.Fatal(err)
	}
	state, _ = masterRestartDecision(opID)
	if state != "succeeded" {
		t.Fatalf("did not reconcile promotion: %s", state)
	}
}

func TestMaintenanceWindowBoundaries(t *testing.T) {
	policy := &models.ServerEntity{UpdateZone: "Asia/Shanghai", UpdateStart: "03:00", UpdateEnd: "04:00"}
	for _, tc := range []struct {
		time string
		want bool
	}{
		{"2026-09-22T18:59:00Z", false},
		{"2026-09-22T19:00:00Z", true},
		{"2026-09-22T19:59:00Z", true},
		{"2026-09-22T20:00:00Z", false},
	} {
		when, err := time.Parse(time.RFC3339, tc.time)
		if err != nil {
			t.Fatal(err)
		}
		if got := withinWindow(policy, when); got != tc.want {
			t.Errorf("withinWindow(%s)=%v, want %v", tc.time, got, tc.want)
		}
	}
	policy.UpdateZone = "Unknown/Zone"
	if withinWindow(policy, time.Now()) {
		t.Fatal("invalid timezone must not schedule an update")
	}
	for _, value := range []string{"24:00", "04:60", "4:00", "04:0a"} {
		if _, ok := timeOfDay(value); ok {
			t.Errorf("accepted invalid time %q", value)
		}
	}
}
