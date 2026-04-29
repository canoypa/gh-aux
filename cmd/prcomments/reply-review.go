package prcomments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newReplyReviewCmd() *cobra.Command {
	var prNumber int
	var commentID int
	var body string

	cmd := &cobra.Command{
		Use:   "reply-review",
		Short: "Reply to a pull request review comment",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			reqBytes, err := json.Marshal(map[string]interface{}{
				"body":        body,
				"in_reply_to": commentID,
			})
			if err != nil {
				return fmt.Errorf("failed to encode request: %w", err)
			}

			var raw rawComment
			path := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
			if err := client.Post(path, bytes.NewReader(reqBytes), &raw); err != nil {
				return fmt.Errorf("failed to post reply: %w", err)
			}

			return writeJSON(os.Stdout, raw.toReviewComment())
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().IntVar(&commentID, "id", 0, "Comment ID to reply to")
	cmd.Flags().StringVar(&body, "body", "", "Reply body text")
	_ = cmd.MarkFlagRequired("pr")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}

