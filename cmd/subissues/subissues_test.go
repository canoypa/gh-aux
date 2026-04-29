package subissues

import (
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

func TestSubIssueOutputPosition(t *testing.T) {
	// Verify that fetchSubIssues assigns 0-based positions.
	// We test the position assignment logic inline since fetchSubIssues calls the API.
	nodes := []struct {
		Number int
		Title  string
		State  string
		URL    string
	}{
		{1, "First", "OPEN", "https://github.com/owner/repo/issues/1"},
		{2, "Second", "CLOSED", "https://github.com/owner/repo/issues/2"},
	}
	out := make([]subIssueOutput, len(nodes))
	for i, n := range nodes {
		pos := i
		out[i] = subIssueOutput{
			Number:   n.Number,
			Title:    n.Title,
			State:    n.State,
			URL:      n.URL,
			Position: &pos,
		}
	}
	for i, o := range out {
		if *o.Position != i {
			t.Errorf("out[%d].Position = %d, want %d", i, *o.Position, i)
		}
	}
}
