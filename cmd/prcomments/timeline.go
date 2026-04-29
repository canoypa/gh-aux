package prcomments

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newTimelineCmd() *cobra.Command {
	var prNumber int
	var cursor string
	var unresolvedOnly bool

	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "List PR comments and review threads in chronological order",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultGraphQLClient()
			if err != nil {
				return fmt.Errorf("failed to create GraphQL client: %w", err)
			}

			// reviewThreads fetches isResolved per thread (keyed by root comment databaseId).
			// timelineItems returns ISSUE_COMMENT and PULL_REQUEST_REVIEW events in order.
			const query = `
query($owner: String!, $repo: String!, $pr: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          isResolved
          comments(first: 1) {
            nodes {
              databaseId
            }
          }
        }
      }
      timelineItems(first: 100, after: $after, itemTypes: [ISSUE_COMMENT, PULL_REQUEST_REVIEW]) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          __typename
          ... on IssueComment {
            databaseId
            body
            author { login }
            createdAt
            url
          }
          ... on PullRequestReview {
            id
            state
            body
            author { login }
            submittedAt
            url
            comments(first: 100) {
              pageInfo {
                hasNextPage
                endCursor
              }
              nodes {
                databaseId
                replyTo { databaseId }
                path
                line
                originalLine
                diffHunk
                body
                author { login }
                createdAt
                url
              }
            }
          }
        }
      }
    }
  }
}`

			type rawReviewComment struct {
				DatabaseID int `json:"databaseId"`
				ReplyTo    *struct {
					DatabaseID int `json:"databaseId"`
				} `json:"replyTo"`
				Path         string `json:"path"`
				Line         int    `json:"line"`
				OriginalLine int    `json:"originalLine"`
				DiffHunk     string `json:"diffHunk"`
				Body         string `json:"body"`
				Author       struct {
					Login string `json:"login"`
				} `json:"author"`
				CreatedAt string `json:"createdAt"`
				URL       string `json:"url"`
			}

			type rawTimelineNode struct {
				Typename string `json:"__typename"`
				// IssueComment
				DatabaseID int    `json:"databaseId"`
				Body       string `json:"body"`
				Author     struct {
					Login string `json:"login"`
				} `json:"author"`
				CreatedAt string `json:"createdAt"`
				URL       string `json:"url"`
				// PullRequestReview
				ID          string `json:"id"`
				State       string `json:"state"`
				SubmittedAt string `json:"submittedAt"`
				Comments    struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []rawReviewComment `json:"nodes"`
				} `json:"comments"`
			}

			var result struct {
				Repository struct {
					PullRequest struct {
						ReviewThreads struct {
							PageInfo struct {
								HasNextPage bool   `json:"hasNextPage"`
								EndCursor   string `json:"endCursor"`
							} `json:"pageInfo"`
							Nodes []struct {
								IsResolved bool `json:"isResolved"`
								Comments   struct {
									Nodes []struct {
										DatabaseID int `json:"databaseId"`
									} `json:"nodes"`
								} `json:"comments"`
							} `json:"nodes"`
						} `json:"reviewThreads"`
						TimelineItems struct {
							PageInfo struct {
								HasNextPage bool   `json:"hasNextPage"`
								EndCursor   string `json:"endCursor"`
							} `json:"pageInfo"`
							Nodes []rawTimelineNode `json:"nodes"`
						} `json:"timelineItems"`
					} `json:"pullRequest"`
				} `json:"repository"`
			}

			var after *string
			if cmd.Flags().Changed("cursor") {
				after = &cursor
			}

			variables := map[string]interface{}{
				"owner": owner,
				"repo":  repo,
				"pr":    prNumber,
				"after": after,
			}

			if err := client.Do(query, variables, &result); err != nil {
				return fmt.Errorf("GraphQL query failed: %w", err)
			}

			// Build isResolved map: root comment databaseId → isResolved.
			// Paginate all reviewThreads pages so the map is always complete.
			isResolvedMap := map[int]bool{}
			for _, thread := range result.Repository.PullRequest.ReviewThreads.Nodes {
				if len(thread.Comments.Nodes) > 0 {
					isResolvedMap[thread.Comments.Nodes[0].DatabaseID] = thread.IsResolved
				}
			}
			if result.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage {
				rtAfter := result.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor
				for {
					var rtResult struct {
						Repository struct {
							PullRequest struct {
								ReviewThreads struct {
									PageInfo struct {
										HasNextPage bool   `json:"hasNextPage"`
										EndCursor   string `json:"endCursor"`
									} `json:"pageInfo"`
									Nodes []struct {
										IsResolved bool `json:"isResolved"`
										Comments   struct {
											Nodes []struct {
												DatabaseID int `json:"databaseId"`
											} `json:"nodes"`
										} `json:"comments"`
									} `json:"nodes"`
								} `json:"reviewThreads"`
							} `json:"pullRequest"`
						} `json:"repository"`
					}
					const rtQuery = `
query($owner: String!, $repo: String!, $pr: Int!, $after: String!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          isResolved
          comments(first: 1) {
            nodes { databaseId }
          }
        }
      }
    }
  }
}`
					if err := client.Do(rtQuery, map[string]interface{}{
						"owner": owner,
						"repo":  repo,
						"pr":    prNumber,
						"after": rtAfter,
					}, &rtResult); err != nil {
						return fmt.Errorf("failed to paginate review threads: %w", err)
					}
					rtPage := rtResult.Repository.PullRequest.ReviewThreads
					for _, thread := range rtPage.Nodes {
						if len(thread.Comments.Nodes) > 0 {
							isResolvedMap[thread.Comments.Nodes[0].DatabaseID] = thread.IsResolved
						}
					}
					if !rtPage.PageInfo.HasNextPage {
						break
					}
					rtAfter = rtPage.PageInfo.EndCursor
				}
			}

			// Paginate inline comments for any review that has > 100 comments.
			const reviewCommentsQuery = `
query($reviewId: ID!, $after: String!) {
  node(id: $reviewId) {
    ... on PullRequestReview {
      comments(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          databaseId
          replyTo { databaseId }
          path
          line
          originalLine
          diffHunk
          body
          author { login }
          createdAt
          url
        }
      }
    }
  }
}`
			for i := range result.Repository.PullRequest.TimelineItems.Nodes {
				node := &result.Repository.PullRequest.TimelineItems.Nodes[i]
				if node.Typename != "PullRequestReview" || !node.Comments.PageInfo.HasNextPage {
					continue
				}
				rcAfter := node.Comments.PageInfo.EndCursor
				for {
					var rcResult struct {
						Node struct {
							Comments struct {
								PageInfo struct {
									HasNextPage bool   `json:"hasNextPage"`
									EndCursor   string `json:"endCursor"`
								} `json:"pageInfo"`
								Nodes []rawReviewComment `json:"nodes"`
							} `json:"comments"`
						} `json:"node"`
					}
					if err := client.Do(reviewCommentsQuery, map[string]interface{}{
						"reviewId": node.ID,
						"after":    rcAfter,
					}, &rcResult); err != nil {
						return fmt.Errorf("failed to paginate comments for review %s: %w", node.ID, err)
					}
					node.Comments.Nodes = append(node.Comments.Nodes, rcResult.Node.Comments.Nodes...)
					if !rcResult.Node.Comments.PageInfo.HasNextPage {
						break
					}
					rcAfter = rcResult.Node.Comments.PageInfo.EndCursor
				}
			}

			// Output types
			type outAuthor struct {
				Login string `json:"login"`
			}

			type outPageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			}

			type outComment struct {
				ID           int       `json:"id"`
				Path         string    `json:"path"`
				Line         int       `json:"line,omitempty"`
				OriginalLine int       `json:"originalLine,omitempty"`
				DiffHunk     string    `json:"diffHunk"`
				Body         string    `json:"body"`
				Author       outAuthor `json:"author"`
				CreatedAt    string    `json:"createdAt"`
				URL          string    `json:"url"`
			}

			type outThread struct {
				IsResolved bool         `json:"isResolved"`
				Comments   []outComment `json:"comments"`
			}

			type outReview struct {
				Type        string      `json:"type"`
				ID          string      `json:"id"`
				State       string      `json:"state"`
				Body        string      `json:"body"`
				Author      outAuthor   `json:"author"`
				SubmittedAt string      `json:"submittedAt"`
				URL         string      `json:"url"`
				Threads     []outThread `json:"threads"`
			}

			type outIssueComment struct {
				Type      string    `json:"type"`
				ID        int       `json:"id"`
				Body      string    `json:"body"`
				Author    outAuthor `json:"author"`
				CreatedAt string    `json:"createdAt"`
				URL       string    `json:"url"`
			}

			type outTimeline struct {
				PageInfo outPageInfo `json:"pageInfo"`
				Nodes    []any       `json:"nodes"`
			}

			tl := result.Repository.PullRequest.TimelineItems
			out := outTimeline{
				PageInfo: outPageInfo{
					HasNextPage: tl.PageInfo.HasNextPage,
					EndCursor:   tl.PageInfo.EndCursor,
				},
				Nodes: make([]any, 0, len(tl.Nodes)),
			}

			for _, node := range tl.Nodes {
				switch node.Typename {
				case "IssueComment":
					out.Nodes = append(out.Nodes, outIssueComment{
						Type:      "comment",
						ID:        node.DatabaseID,
						Body:      truncateBody(node.Body),
						Author:    outAuthor{Login: node.Author.Login},
						CreatedAt: node.CreatedAt,
						URL:       node.URL,
					})

				case "PullRequestReview":
					// Group inline comments into threads using replyTo as the root indicator.
					threadMap := map[int]*outThread{}
					threadOrder := []int{}

					for _, c := range node.Comments.Nodes {
						rootID := c.DatabaseID
						if c.ReplyTo != nil {
							rootID = c.ReplyTo.DatabaseID
						}

						if _, exists := threadMap[rootID]; !exists {
							resolved := isResolvedMap[rootID]
							if unresolvedOnly && resolved {
								continue
							}
							threadMap[rootID] = &outThread{IsResolved: resolved}
							threadOrder = append(threadOrder, rootID)
						} else if unresolvedOnly && threadMap[rootID].IsResolved {
							continue
						}

						threadMap[rootID].Comments = append(threadMap[rootID].Comments, outComment{
							ID:           c.DatabaseID,
							Path:         c.Path,
							Line:         c.Line,
							OriginalLine: c.OriginalLine,
							DiffHunk:     c.DiffHunk,
							Body:         truncateBody(c.Body),
							Author:       outAuthor{Login: c.Author.Login},
							CreatedAt:    c.CreatedAt,
							URL:          c.URL,
						})
					}

					threads := make([]outThread, 0, len(threadOrder))
					for _, rootID := range threadOrder {
						threads = append(threads, *threadMap[rootID])
					}

					// Skip reviews with no matching threads when filtering
					if unresolvedOnly && len(threads) == 0 {
						continue
					}

					out.Nodes = append(out.Nodes, outReview{
						Type:        "review",
						ID:          node.ID,
						State:       node.State,
						Body:        truncateBody(node.Body),
						Author:      outAuthor{Login: node.Author.Login},
						SubmittedAt: node.SubmittedAt,
						URL:         node.URL,
						Threads:     threads,
					})
				}
			}

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().StringVar(&cursor, "cursor", "", "Pagination cursor from a previous response")
	cmd.Flags().BoolVar(&unresolvedOnly, "unresolved", false, "Only return unresolved review threads")
	_ = cmd.MarkFlagRequired("pr")

	return cmd
}
