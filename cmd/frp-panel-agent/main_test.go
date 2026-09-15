package main

import "testing"

func TestNodeIDMatchesEnrollment(t *testing.T) {
	tests := []struct {
		requested, enrolled string
		want                bool
	}{
		{"owner.c.mac", "owner.c.mac", true},
		{"mac", "owner.c.mac", true},
		{"", "owner.c.mac", true},
		{"other", "owner.c.mac", false},
	}
	for _, test := range tests {
		if got := nodeIDMatchesEnrollment(test.requested, test.enrolled); got != test.want {
			t.Fatalf("nodeIDMatchesEnrollment(%q, %q) = %t, want %t", test.requested, test.enrolled, got, test.want)
		}
	}
}
