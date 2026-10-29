package reviewcomments

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newResolveThreadCmd() *cobra.Command {
	var threadID string
	var commentID int

	cmd := &cobra.Command{
		Use:   "resolve-thread",
		Short: "Resolve a pull request review thread",
		RunE: func(cmd *cobra.Command, args []string) error {
			hasThreadID := cmd.Flags().Changed("thread-id")
			hasCommentID := cmd.Flags().Changed("comment-id")

			if !hasThreadID && !hasCommentID {
				return fmt.Errorf("one of --thread-id or --comment-id is required")
			}
			if hasThreadID && hasCommentID {
				return fmt.Errorf("--thread-id and --comment-id are mutually exclusive")
			}

			client, err := api.DefaultGraphQLClient()
			if err != nil {
				return fmt.Errorf("failed to create GraphQL client: %w", err)
			}

			resolvedThreadID := threadID

			if hasCommentID {
				if commentID <= 0 {
					return fmt.Errorf("comment-id must be > 0")
				}
				owner, repo, err := resolveRepo(repoFlag)
				if err != nil {
					return err
				}
				restClient, err := api.DefaultRESTClient()
				if err != nil {
					return fmt.Errorf("failed to create REST client: %w", err)
				}
				var commentResp struct {
					PullRequestURL string `json:"pull_request_url"`
				}
				path := fmt.Sprintf("repos/%s/%s/pulls/comments/%d", owner, repo, commentID)
				if err := restClient.Get(path, &commentResp); err != nil {
					return fmt.Errorf("failed to fetch comment: %w", err)
				}
				prNumber, err := parsePullRequestNumber(commentResp.PullRequestURL)
				if err != nil {
					return err
				}
				resolvedThreadID, err = findThreadNodeID(client, owner, repo, prNumber, commentID)
				if err != nil {
					return err
				}
			}

			const mutation = `
mutation($threadId: ID!) {
  resolveReviewThread(input: { threadId: $threadId }) {
    thread {
      id
      isResolved
    }
  }
}`
			type resolveResp struct {
				ResolveReviewThread struct {
					Thread struct {
						ID         string `json:"id"`
						IsResolved bool   `json:"isResolved"`
					} `json:"thread"`
				} `json:"resolveReviewThread"`
			}

			var result resolveResp
			if err := client.Do(mutation, map[string]interface{}{
				"threadId": resolvedThreadID,
			}, &result); err != nil {
				return fmt.Errorf("failed to resolve thread: %w", err)
			}

			type output struct {
				ThreadID   string `json:"threadId"`
				IsResolved bool   `json:"isResolved"`
			}
			return writeJSON(os.Stdout, output{
				ThreadID:   result.ResolveReviewThread.Thread.ID,
				IsResolved: result.ResolveReviewThread.Thread.IsResolved,
			})
		},
	}

	cmd.Flags().StringVar(&threadID, "thread-id", "", "Thread node ID to resolve")
	cmd.Flags().IntVar(&commentID, "comment-id", 0, "Review comment integer ID from URL (#discussion_r<id>); requires --repo")

	return cmd
}
