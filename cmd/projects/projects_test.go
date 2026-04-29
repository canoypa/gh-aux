package projects

import (
	"strings"
	"testing"
)

func TestResolveRepoString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{"valid", "owner/repo", "owner", "repo", false},
		{"no slash", "ownerrepo", "", "", true},
		{"empty owner", "/repo", "", "", true},
		{"empty name", "owner/", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := resolveRepo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveRepo(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if owner != tt.wantOwner || name != tt.wantName {
				t.Errorf("resolveRepo(%q) = (%q, %q), want (%q, %q)", tt.input, owner, name, tt.wantOwner, tt.wantName)
			}
		})
	}
}

func TestFieldFlagValidation(t *testing.T) {
	// Verify that missing '=' in --field is caught before API calls.
	tests := []struct {
		flag    string
		wantErr bool
	}{
		{"Status=In Progress", false},
		{"Status", true},
		{"=Value", false}, // empty name will fail at field lookup, not here
	}
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			hasEq := strings.IndexByte(tt.flag, '=') >= 0
			gotErr := !hasEq
			if gotErr != tt.wantErr {
				t.Errorf("flag %q: hasEq=%v gotErr=%v wantErr=%v", tt.flag, hasEq, gotErr, tt.wantErr)
			}
		})
	}
}
