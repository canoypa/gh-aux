package prcomments

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
				// Step 1: REST GET to obtain the comment's GraphQL node ID.
				owner, repo, err := resolveRepo(repoFlag)
				if err != nil {
					return err
				}
				restClient, err := api.DefaultRESTClient()
				if err != nil {
					return fmt.Errorf("failed to create REST client: %w", err)
				}
				var commentResp struct {
					NodeID string `json:"node_id"`
				}
				path := fmt.Sprintf("repos/%s/%s/pulls/comments/%d", owner, repo, commentID)
				if err := restClient.Get(path, &commentResp); err != nil {
					return fmt.Errorf("failed to fetch comment: %w", err)
				}

				// Step 2: GraphQL to resolve comment node → thread node ID.
				const queryThread = `
query($nodeId: ID!) {
  node(id: $nodeId) {
    ... on PullRequestReviewComment {
      thread { id }
    }
  }
}`
				var threadResp struct {
					Node struct {
						Thread struct {
							ID string `json:"id"`
						} `json:"thread"`
					} `json:"node"`
				}
				if err := client.Do(queryThread, map[string]interface{}{
					"nodeId": commentResp.NodeID,
				}, &threadResp); err != nil {
					return fmt.Errorf("failed to resolve thread from comment: %w", err)
				}
				resolvedThreadID = threadResp.Node.Thread.ID
				if resolvedThreadID == "" {
					return fmt.Errorf("could not resolve thread for comment ID %d", commentID)
				}
			}

			// Resolve the thread.
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

	cmd.Flags().StringVar(&threadID, "thread-id", "", "Thread node ID to resolve (from timeline threads[].id)")
	cmd.Flags().IntVar(&commentID, "comment-id", 0, "Review comment integer ID from URL (#discussion_r<id>); requires --repo")

	return cmd
}
