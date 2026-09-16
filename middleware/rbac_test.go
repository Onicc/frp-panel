package middleware

import (
	"testing"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
)

func TestRolePermissionsUseCurrentStoredRoleRules(t *testing.T) {
	tests := []struct {
		role, method, path string
		want               bool
	}{
		{defs.UserRole_Viewer, "POST", "/api/v1/client/list", true},
		{defs.UserRole_Viewer, "POST", "/api/v1/client/delete", false},
		{defs.UserRole_Operator, "POST", "/api/v2/enrollments", true},
		{defs.UserRole_Operator, "POST", "/api/v2/server-enrollments", true},
		{defs.UserRole_Operator, "POST", "/api/v2/tunnels", true},
		{defs.UserRole_Operator, "DELETE", "/api/v2/tunnels", true},
		{defs.UserRole_Viewer, "POST", "/api/v2/account/password", true},
		{defs.UserRole_Operator, "POST", "/api/v1/user/admin-update", false},
		{defs.UserRole_Admin, "DELETE", "/api/v2/anything", true},
	}
	for _, test := range tests {
		allowed := false
		for _, permission := range conf.PermissionsForRole(test.role) {
			if ruleMatched(ruleMatchParam{
				RuleMethod: permission.Method, RulePath: permission.Path,
				RequestMethod: test.method, RequestPath: test.path,
			}) {
				allowed = true
				break
			}
		}
		if allowed != test.want {
			t.Fatalf("role %s %s %s allowed=%t, want %t", test.role, test.method, test.path, allowed, test.want)
		}
	}
}
