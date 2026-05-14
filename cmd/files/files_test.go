package files

import (
	"bytes"
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
		{"extra slash", "owner/repo/extra", "", "", true},
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

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeJSON(&buf, map[string]int{"x": 1}); err != nil {
		t.Fatalf("writeJSON error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"x": 1`) {
		t.Errorf("writeJSON output %q does not contain expected content", out)
	}
}

func TestWriteJSONNil(t *testing.T) {
	var buf bytes.Buffer
	if err := writeJSON(&buf, nil); err != nil {
		t.Fatalf("writeJSON(nil) error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "null" {
		t.Errorf("writeJSON(nil) = %q, want %q", buf.String(), "null")
	}
}
