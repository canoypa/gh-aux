package files

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

// rawContents is the GitHub REST API response for a file at /repos/{owner}/{repo}/contents/{path}.
type rawContents struct {
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Content string `json:"content"`
	HTMLURL string `json:"html_url"`
}

// rawDirEntry is one item in the GitHub REST API directory listing response.
type rawDirEntry struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Size    int    `json:"size"`
	HTMLURL string `json:"html_url"`
}

// DirEntry is the normalized output for a single directory entry.
type DirEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

func newGetCmd() *cobra.Command {
	var filePath string
	var ref string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get file contents or directory listing from a repository at a given ref",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			trimmedPath := strings.TrimPrefix(filePath, "/")
			pathSegments := strings.Split(trimmedPath, "/")
			for i, seg := range pathSegments {
				pathSegments[i] = url.PathEscape(seg)
			}
			apiPath := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, strings.Join(pathSegments, "/"))
			if ref != "" {
				apiPath += "?ref=" + url.QueryEscape(ref)
			}

			var raw json.RawMessage
			if err := client.Get(apiPath, &raw); err != nil {
				return fmt.Errorf("failed to get %q: %w", filePath, err)
			}

			resolvedRef := ref
			if resolvedRef == "" {
				resolvedRef = "HEAD"
			}

			// The API returns an array for directories and an object for files.
			if len(raw) > 0 && raw[0] == '[' {
				var entries []rawDirEntry
				if err := json.Unmarshal(raw, &entries); err != nil {
					return fmt.Errorf("failed to parse directory listing: %w", err)
				}
				out := make([]DirEntry, len(entries))
				for i, e := range entries {
					out[i] = DirEntry{Type: e.Type, Name: e.Name, Path: e.Path, SHA: e.SHA, Size: e.Size, URL: e.HTMLURL}
				}
				return writeJSON(os.Stdout, out)
			}

			var file rawContents
			if err := json.Unmarshal(raw, &file); err != nil {
				return fmt.Errorf("failed to parse file contents: %w", err)
			}

			// GitHub returns base64-encoded content with embedded newlines.
			decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(file.Content, "\n", ""))
			if err != nil {
				return fmt.Errorf("failed to decode content: %w", err)
			}

			return writeJSON(os.Stdout, FileContent{
				Path:    file.Path,
				Ref:     resolvedRef,
				SHA:     file.SHA,
				Content: string(decoded),
				URL:     file.HTMLURL,
			})
		},
	}

	cmd.Flags().StringVar(&filePath, "path", "", "File path within the repository (e.g. src/foo.ts)")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag, or commit SHA (defaults to the repository's default branch)")
	_ = cmd.MarkFlagRequired("path")

	return cmd
}
