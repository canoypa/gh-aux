package prcomments

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)


// ReviewComment represents a pull request review comment (REST API shape, normalized to camelCase).
type ReviewComment struct {
	ID           int    `json:"id"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	Line         int    `json:"line"`
	OriginalLine int    `json:"originalLine"`
	Author       struct {
		Login string `json:"login"`
	} `json:"author"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	URL       string `json:"url"`
}

const bodyTruncateLen = 200

// truncateBody truncates s to bodyTruncateLen runes, appending "..." if truncated.
func truncateBody(s string) string {
	runes := []rune(s)
	if len(runes) <= bodyTruncateLen {
		return s
	}
	return string(runes[:bodyTruncateLen]) + "..."
}

// rawComment is the REST API response shape for a pull request review comment.
type rawComment struct {
	ID           int    `json:"id"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	Line         *int   `json:"line"`
	OriginalLine *int   `json:"original_line"`
	User         struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
}

// toReviewComment converts a rawComment to a ReviewComment.
func (r rawComment) toReviewComment() ReviewComment {
	c := ReviewComment{
		ID:        r.ID,
		Body:      r.Body,
		Path:      r.Path,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		URL:       r.HTMLURL,
	}
	c.Author.Login = r.User.Login
	if r.Line != nil {
		c.Line = *r.Line
	}
	if r.OriginalLine != nil {
		c.OriginalLine = *r.OriginalLine
	}
	return c
}

// resolveRepo resolves owner and repository name from a "OWNER/REPO" string,
// falling back to the current directory's git remote if the argument is empty.
func resolveRepo(repoStr string) (owner, name string, err error) {
	if repoStr != "" {
		parts := strings.SplitN(repoStr, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid repo format %q: expected OWNER/REPO", repoStr)
		}
		return parts[0], parts[1], nil
	}
	r, err := repository.Current()
	if err != nil {
		return "", "", fmt.Errorf("could not determine current repository (use --repo): %w", err)
	}
	return r.Owner, r.Name, nil
}

// IssueComment represents a general (non-diff) pull request comment.
type IssueComment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	URL       string `json:"url"`
}

// rawIssueComment is the REST API response shape for an issue/PR comment.
type rawIssueComment struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
}

func (r rawIssueComment) toIssueComment() IssueComment {
	c := IssueComment{
		ID:        r.ID,
		Body:      r.Body,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		URL:       r.HTMLURL,
	}
	c.Author.Login = r.User.Login
	return c
}

// writeJSON writes v as indented JSON to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
