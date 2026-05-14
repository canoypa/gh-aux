package files

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

// DownloadResult is the output of the download command.
type DownloadResult struct {
	Path   string `json:"path"`
	Output string `json:"output"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
	URL    string `json:"url"`
}

func newDownloadCmd() *cobra.Command {
	var filePath string
	var ref string
	var output string

	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a file from a repository to a local path",
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

			var rawMsg json.RawMessage
			if err := client.Get(apiPath, &rawMsg); err != nil {
				return fmt.Errorf("failed to get file %q: %w", filePath, err)
			}
			if len(rawMsg) > 0 && rawMsg[0] == '[' {
				return fmt.Errorf("%q is a directory; use 'files get' to list directory contents", filePath)
			}
			var raw rawContents
			if err := json.Unmarshal(rawMsg, &raw); err != nil {
				return fmt.Errorf("failed to parse file contents: %w", err)
			}

			decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(raw.Content, "\n", ""))
			if err != nil {
				return fmt.Errorf("failed to decode content: %w", err)
			}

			if err := os.WriteFile(output, decoded, 0644); err != nil {
				return fmt.Errorf("failed to write file %q: %w", output, err)
			}

			// Resolve output to an absolute path for clarity.
			abs, err := filepath.Abs(output)
			if err != nil {
				abs = output
			}

			resolvedRef := ref
			if resolvedRef == "" {
				resolvedRef = "HEAD"
			}

			return writeJSON(os.Stdout, DownloadResult{
				Path:   raw.Path,
				Output: abs,
				Ref:    resolvedRef,
				SHA:    raw.SHA,
				URL:    raw.HTMLURL,
			})
		},
	}

	cmd.Flags().StringVar(&filePath, "path", "", "File path within the repository (e.g. src/foo.ts)")
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag, or commit SHA (defaults to the repository's default branch)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Local path to write the file to")
	_ = cmd.MarkFlagRequired("path")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}
