package files

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// FileContent is the normalized output for a file fetched from GitHub.
type FileContent struct {
	Path    string `json:"path"`
	Ref     string `json:"ref"`
	SHA     string `json:"sha"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// resolveRepo resolves owner and repository name from a "OWNER/REPO" string,
// falling back to the current directory's git remote if the argument is empty.
func resolveRepo(repoStr string) (owner, name string, err error) {
	if repoStr != "" {
		parts := strings.Split(repoStr, "/")
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

// writeJSON encodes v as JSON and writes it to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
