package prcomments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var prNumber int
	var body string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a general comment to a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			reqBytes, err := json.Marshal(map[string]string{
				"body": body,
			})
			if err != nil {
				return fmt.Errorf("failed to encode request body: %w", err)
			}

			var raw rawIssueComment
			path := fmt.Sprintf("repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
			if err := client.Post(path, bytes.NewReader(reqBytes), &raw); err != nil {
				return fmt.Errorf("failed to add comment: %w", err)
			}

			return writeJSON(os.Stdout, raw.toIssueComment())
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	_ = cmd.MarkFlagRequired("pr")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}
