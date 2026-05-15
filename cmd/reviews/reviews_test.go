package reviews

import (
	"strings"
	"testing"
)

func TestResolveRepo(t *testing.T) {
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

func TestValidateComments(t *testing.T) {
	tests := []struct {
		name    string
		input   []reviewCommentInput
		wantErr bool
		errMsg  string
	}{
		{"empty slice", []reviewCommentInput{}, false, ""},
		{"valid reply", []reviewCommentInput{{InReplyTo: 1, Body: "reply"}}, false, ""},
		{"valid file-level", []reviewCommentInput{{Path: "foo.go", Body: "note", SubjectType: "file"}}, false, ""},
		{"valid line-level", []reviewCommentInput{{Path: "foo.go", Body: "note", Line: 10, Side: "RIGHT"}}, false, ""},
		{"valid multi-line", []reviewCommentInput{{Path: "foo.go", Body: "note", Line: 10, Side: "RIGHT", StartLine: 8, StartSide: "RIGHT"}}, false, ""},
		{"reply missing body", []reviewCommentInput{{InReplyTo: 1, Body: ""}}, true, "body is required for replies"},
		{"missing path", []reviewCommentInput{{Body: "note", Line: 1, Side: "RIGHT"}}, true, "path is required"},
		{"missing body", []reviewCommentInput{{Path: "foo.go", Line: 10, Side: "RIGHT"}}, true, "body is required"},
		{"line zero", []reviewCommentInput{{Path: "foo.go", Body: "b", Line: 0, Side: "RIGHT"}}, true, "line must be > 0"},
		{"invalid side", []reviewCommentInput{{Path: "foo.go", Body: "b", Line: 1, Side: "right"}}, true, "side must be"},
		{"start_line > line", []reviewCommentInput{{Path: "foo.go", Body: "b", Line: 5, Side: "RIGHT", StartLine: 10, StartSide: "RIGHT"}}, true, "start_line"},
		{"start_side missing", []reviewCommentInput{{Path: "foo.go", Body: "b", Line: 10, Side: "RIGHT", StartLine: 8}}, true, "start_side is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComments(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateComments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateComments() error = %q, want to contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}
