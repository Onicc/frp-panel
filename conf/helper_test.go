package conf

import (
	"testing"

	"github.com/Onicc/frp-panel/defs"
)

func TestPermissionsForRole(t *testing.T) {
	owner := PermissionsForRole(defs.UserRole_Owner)
	if len(owner) != 1 || owner[0].Method != "*" || owner[0].Path != "*" {
		t.Fatalf("owner permissions: %#v", owner)
	}
	viewer := PermissionsForRole(defs.UserRole_Viewer)
	if len(viewer) < 2 {
		t.Fatalf("viewer cannot access read endpoints: %#v", viewer)
	}
	operator := PermissionsForRole(defs.UserRole_Operator)
	if len(operator) <= len(viewer) {
		t.Fatalf("operator lacks mutation permissions: %#v", operator)
	}
}
