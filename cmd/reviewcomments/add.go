package reviewcomments

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
	var inReplyTo int
	var path string
	var line int
	var side string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an inline review comment to a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if prNumber <= 0 {
				return fmt.Errorf("pr must be > 0")
			}

			hasReplyTo := cmd.Flags().Changed("in-reply-to")

			if hasReplyTo && (path != "" || line != 0 || side != "") {
				return fmt.Errorf("--in-reply-to is mutually exclusive with --path, --line, and --side")
			}

			if !hasReplyTo {
				if path == "" {
					return fmt.Errorf("--path is required when --in-reply-to is not set")
				}
				if line <= 0 {
					return fmt.Errorf("--line must be > 0")
				}
				if side != "LEFT" && side != "RIGHT" {
					return fmt.Errorf("--side must be LEFT or RIGHT (uppercase)")
				}
			}

			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			var reqBody map[string]interface{}
			if hasReplyTo {
				if inReplyTo <= 0 {
					return fmt.Errorf("--in-reply-to must be > 0")
				}
				reqBody = map[string]interface{}{
					"body":        body,
					"in_reply_to": inReplyTo,
				}
			} else {
				// Fetch head commit SHA for new standalone comment.
				var pr struct {
					Head struct {
						SHA string `json:"sha"`
					} `json:"head"`
				}
				prPath := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, prNumber)
				if err := client.Get(prPath, &pr); err != nil {
					return fmt.Errorf("failed to get PR: %w", err)
				}
				reqBody = map[string]interface{}{
					"body":      body,
					"commit_id": pr.Head.SHA,
					"path":      path,
					"line":      line,
					"side":      side,
				}
			}

			reqBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to encode request: %w", err)
			}

			var raw rawComment
			apiPath := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
			if err := client.Post(apiPath, bytes.NewReader(reqBytes), &raw); err != nil {
				return fmt.Errorf("failed to add comment: %w", err)
			}

			return writeJSON(os.Stdout, raw.toReviewComment())
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().StringVar(&body, "body", "", "Comment body text")
	cmd.Flags().IntVar(&inReplyTo, "in-reply-to", 0, "Comment ID to reply to (makes --path/--line/--side optional)")
	cmd.Flags().StringVar(&path, "path", "", "File path (required when not replying)")
	cmd.Flags().IntVar(&line, "line", 0, "Line number (required when not replying)")
	cmd.Flags().StringVar(&side, "side", "", "Side: LEFT or RIGHT (required when not replying)")
	_ = cmd.MarkFlagRequired("pr")
	_ = cmd.MarkFlagRequired("body")

	return cmd
}
