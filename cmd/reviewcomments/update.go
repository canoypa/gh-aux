package reviewcomments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var commentID int
	var body string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the body of an inline review comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commentID <= 0 {
				return fmt.Errorf("id must be > 0")
			}
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			reqBytes, err := json.Marshal(map[string]string{"body": body})
			if err != nil {
				return fmt.Errorf("failed to encode request body: %w", err)
			}

			var raw rawComment
			path := fmt.Sprintf("repos/%s/%s/pulls/comments/%d", owner, repo, commentID)
			if err := client.Patch(path, bytes.NewReader(reqBytes), &raw); err != nil {
				return fmt.Errorf("failed to update comment %d: %w", commentID, err)
			}

			return writeJSON(os.Stdout, raw.toReviewComment())
		},
	}

	cmd.Flags().IntVar(&commentID, "id", 0, "Inline review comment ID to update")
	cmd.Flags().StringVar(&body, "body", "", "New comment body text")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}
