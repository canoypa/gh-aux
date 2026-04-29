package prcomments

import (
	"bytes"
	"strings"
	"testing"
)

func TestTruncateBody(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"short", "hello", "hello"},
		{"exactly 200 runes", strings.Repeat("a", 200), strings.Repeat("a", 200)},
		{"201 runes gets truncated", strings.Repeat("a", 201), strings.Repeat("a", 200) + "..."},
		{"multibyte 200 runes", strings.Repeat("あ", 200), strings.Repeat("あ", 200)},
		{"multibyte 201 runes", strings.Repeat("あ", 201), strings.Repeat("あ", 200) + "..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateBody(tt.input)
			if got != tt.want {
				t.Errorf("truncateBody(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

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
		{"extra slash is valid", "owner/repo/extra", "owner", "repo/extra", false},
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

func TestRawCommentToReviewComment(t *testing.T) {
	line := 42
	originalLine := 40
	raw := rawComment{
		ID:           123,
		Body:         "test body",
		Path:         "src/foo.go",
		Line:         &line,
		OriginalLine: &originalLine,
		User:         struct{ Login string `json:"login"` }{Login: "alice"},
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-02T00:00:00Z",
		HTMLURL:      "https://github.com/owner/repo/pull/1#discussion_r123",
	}

	got := raw.toReviewComment()

	if got.ID != 123 {
		t.Errorf("ID = %d, want 123", got.ID)
	}
	if got.Body != "test body" {
		t.Errorf("Body = %q, want %q", got.Body, "test body")
	}
	if got.Line != 42 {
		t.Errorf("Line = %d, want 42", got.Line)
	}
	if got.OriginalLine != 40 {
		t.Errorf("OriginalLine = %d, want 40", got.OriginalLine)
	}
	if got.Author.Login != "alice" {
		t.Errorf("Author.Login = %q, want %q", got.Author.Login, "alice")
	}
	if got.URL != "https://github.com/owner/repo/pull/1#discussion_r123" {
		t.Errorf("URL = %q", got.URL)
	}
}

func TestRawCommentNilLines(t *testing.T) {
	raw := rawComment{ID: 1, Body: "b"}
	got := raw.toReviewComment()
	if got.Line != 0 {
		t.Errorf("Line should be 0 when nil, got %d", got.Line)
	}
	if got.OriginalLine != 0 {
		t.Errorf("OriginalLine should be 0 when nil, got %d", got.OriginalLine)
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
