package client

import (
	"reflect"
	"sort"
	"testing"
)

func TestExpiredLegacyClientIDsKeepsPhysicalClient(t *testing.T) {
	tests := []struct {
		name          string
		configuredIDs []string
		controllerIDs []string
		expectedIDs   []string
	}{
		{
			name:          "v2 controller has no legacy children",
			configuredIDs: nil,
			controllerIDs: []string{"onicc.c.plac"},
			expectedIDs:   nil,
		},
		{
			name:          "physical client is never treated as a child",
			configuredIDs: []string{"onicc.c.plac"},
			controllerIDs: []string{"onicc.c.plac"},
			expectedIDs:   nil,
		},
		{
			name:          "legacy child is retained",
			configuredIDs: []string{"shadow-1"},
			controllerIDs: []string{"onicc.c.plac", "shadow-1"},
			expectedIDs:   nil,
		},
		{
			name:          "only expired legacy children are removed",
			configuredIDs: []string{"shadow-1"},
			controllerIDs: []string{"onicc.c.plac", "shadow-1", "shadow-2"},
			expectedIDs:   []string{"shadow-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expiredLegacyClientIDs("onicc.c.plac", tt.configuredIDs, tt.controllerIDs)
			sort.Strings(got)
			sort.Strings(tt.expectedIDs)
			if len(got) != len(tt.expectedIDs) || (len(got) > 0 && !reflect.DeepEqual(got, tt.expectedIDs)) {
				t.Fatalf("expiredLegacyClientIDs() = %v, want %v", got, tt.expectedIDs)
			}
		})
	}
}
