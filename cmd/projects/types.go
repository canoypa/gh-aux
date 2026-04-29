package projects

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// projectItemOutput is the normalized output for a project item.
type projectItemOutput struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Content   struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		URL    string `json:"url"`
	} `json:"content"`
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

// writeJSON serializes v as JSON to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
