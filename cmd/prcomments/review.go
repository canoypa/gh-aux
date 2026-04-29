package prcomments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

// reviewCommentInput is the input shape for a single review comment.
type reviewCommentInput struct {
	Path      string `json:"path,omitempty"`
	Line      int    `json:"line,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	Side      string `json:"side,omitempty"`
	Body      string `json:"body"`
	InReplyTo int    `json:"in_reply_to,omitempty"`
}

// reviewInput is the input schema accepted by the review command.
type reviewInput struct {
	Event    string               `json:"event,omitempty"`
	Body     string               `json:"body,omitempty"`
	Comments []reviewCommentInput `json:"comments,omitempty"`
}

// rawReviewResponse is the REST API response shape for a pull request review.
type rawReviewResponse struct {
	ID     int    `json:"id"`
	NodeID string `json:"node_id"`
	Body   string `json:"body"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
	} `json:"user"`
	HTMLURL     string `json:"html_url"`
	SubmittedAt string `json:"submitted_at"`
}

// reviewOutput is the normalized output for a review.
type reviewOutput struct {
	ID     int    `json:"id"`
	State  string `json:"state"`
	Body   string `json:"body"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	URL         string `json:"url"`
	SubmittedAt string `json:"submittedAt,omitempty"`
}

func (r rawReviewResponse) toOutput() reviewOutput {
	out := reviewOutput{
		ID:          r.ID,
		State:       r.State,
		Body:        r.Body,
		URL:         r.HTMLURL,
		SubmittedAt: r.SubmittedAt,
	}
	out.Author.Login = r.User.Login
	return out
}

// findOrCreatePendingReview returns the current user's pending review,
// creating one (with the PR's head commit) if none exists.
func findOrCreatePendingReview(restClient *api.RESTClient, owner, repo string, prNumber int) (rawReviewResponse, error) {
	var currentUser struct {
		Login string `json:"login"`
	}
	if err := restClient.Get("user", &currentUser); err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to get current user: %w", err)
	}

	reviewsPath := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", owner, repo, prNumber)

	var reviews []rawReviewResponse
	if err := restClient.Get(reviewsPath, &reviews); err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to get reviews: %w", err)
	}

	for _, r := range reviews {
		if r.State == "PENDING" && r.User.Login == currentUser.Login {
			return r, nil
		}
	}

	// No pending review — get head SHA and create one.
	var pr struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := restClient.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, prNumber), &pr); err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to get PR: %w", err)
	}

	reqBytes, err := json.Marshal(map[string]string{"commit_id": pr.Head.SHA})
	if err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to encode request: %w", err)
	}

	var created rawReviewResponse
	if err := restClient.Post(reviewsPath, bytes.NewReader(reqBytes), &created); err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to create pending review: %w", err)
	}

	return created, nil
}

// findThreadNodeID returns the GraphQL node_id of the review thread that contains commentID.
func findThreadNodeID(gqlClient *api.GraphQLClient, owner, repo string, prNumber, commentID int) (string, error) {
	const query = `
query($owner: String!, $repo: String!, $pr: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100) {
        nodes {
          id
          comments(first: 100) {
            nodes { databaseId }
          }
        }
      }
    }
  }
}`
	var result struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						ID       string `json:"id"`
						Comments struct {
							Nodes []struct {
								DatabaseID int `json:"databaseId"`
							} `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}

	if err := gqlClient.Do(query, map[string]interface{}{
		"owner": owner,
		"repo":  repo,
		"pr":    prNumber,
	}, &result); err != nil {
		return "", fmt.Errorf("failed to query threads: %w", err)
	}

	for _, thread := range result.Repository.PullRequest.ReviewThreads.Nodes {
		for _, c := range thread.Comments.Nodes {
			if c.DatabaseID == commentID {
				return thread.ID, nil
			}
		}
	}

	return "", fmt.Errorf("thread containing comment %d not found", commentID)
}

// addCommentToReview adds a single inline comment (new thread or reply) to a pending review via GraphQL.
func addCommentToReview(gqlClient *api.GraphQLClient, owner, repo string, prNumber int, reviewNodeID string, c reviewCommentInput) error {
	if c.InReplyTo != 0 {
		threadNodeID, err := findThreadNodeID(gqlClient, owner, repo, prNumber, c.InReplyTo)
		if err != nil {
			return err
		}

		const mutation = `
mutation($input: AddPullRequestReviewThreadReplyInput!) {
  addPullRequestReviewThreadReply(input: $input) {
    comment { databaseId }
  }
}`
		var result struct {
			AddPullRequestReviewThreadReply struct {
				Comment struct {
					DatabaseID int `json:"databaseId"`
				} `json:"comment"`
			} `json:"addPullRequestReviewThreadReply"`
		}
		if err := gqlClient.Do(mutation, map[string]interface{}{
			"input": map[string]interface{}{
				"pullRequestReviewId":       reviewNodeID,
				"pullRequestReviewThreadId": threadNodeID,
				"body":                      c.Body,
			},
		}, &result); err != nil {
			return fmt.Errorf("failed to add reply to comment %d: %w", c.InReplyTo, err)
		}
		return nil
	}

	// New inline thread.
	side := c.Side
	if side == "" {
		side = "RIGHT"
	}
	input := map[string]interface{}{
		"pullRequestReviewId": reviewNodeID,
		"path":                c.Path,
		"line":                c.Line,
		"side":                side,
		"body":                c.Body,
	}
	if c.StartLine != 0 {
		input["startLine"] = c.StartLine
	}

	const mutation = `
mutation($input: AddPullRequestReviewThreadInput!) {
  addPullRequestReviewThread(input: $input) {
    thread { id }
  }
}`
	var result struct {
		AddPullRequestReviewThread struct {
			Thread struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"addPullRequestReviewThread"`
	}
	if err := gqlClient.Do(mutation, map[string]interface{}{"input": input}, &result); err != nil {
		return fmt.Errorf("failed to add comment to %s:%d: %w", c.Path, c.Line, err)
	}
	return nil
}

// submitPendingReview submits a pending review via REST.
func submitPendingReview(restClient *api.RESTClient, owner, repo string, prNumber, reviewID int, event, body string) (rawReviewResponse, error) {
	reqBytes, err := json.Marshal(map[string]string{
		"event": event,
		"body":  body,
	})
	if err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to encode request: %w", err)
	}

	var result rawReviewResponse
	path := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews/%d/events", owner, repo, prNumber, reviewID)
	if err := restClient.Post(path, bytes.NewReader(reqBytes), &result); err != nil {
		return rawReviewResponse{}, fmt.Errorf("failed to submit review: %w", err)
	}
	return result, nil
}

// executeReview is the shared logic for the review and reply-review commands.
func executeReview(owner, repo string, prNumber int, input reviewInput) (reviewOutput, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return reviewOutput{}, fmt.Errorf("failed to create REST client: %w", err)
	}

	// Step 1: find or create pending review.
	pending, err := findOrCreatePendingReview(restClient, owner, repo, prNumber)
	if err != nil {
		return reviewOutput{}, err
	}

	// Step 2: add inline comments via GraphQL.
	if len(input.Comments) > 0 {
		gqlClient, err := api.DefaultGraphQLClient()
		if err != nil {
			return reviewOutput{}, fmt.Errorf("failed to create GraphQL client: %w", err)
		}
		for _, c := range input.Comments {
			if err := addCommentToReview(gqlClient, owner, repo, prNumber, pending.NodeID, c); err != nil {
				return reviewOutput{}, err
			}
		}
	}

	// Step 3: submit if event provided, otherwise leave as pending.
	if input.Event != "" {
		result, err := submitPendingReview(restClient, owner, repo, prNumber, pending.ID, input.Event, input.Body)
		if err != nil {
			return reviewOutput{}, err
		}
		return result.toOutput(), nil
	}

	return pending.toOutput(), nil
}

func newReviewCmd() *cobra.Command {
	var prNumber int
	var jsonStr string

	cmd := &cobra.Command{
		Use:   "review",
		Short: "Create or submit a pull request review",
		Long: `Create or submit a pull request review.

Accepts JSON via --json flag or stdin.

Schema:
  {
    "event":    "APPROVE | REQUEST_CHANGES | COMMENT",  // omit to leave as pending
    "body":     "overall review comment",
    "comments": [
      { "path": "src/foo.ts", "line": 42, "body": "...", "side": "RIGHT" },
      { "path": "src/bar.ts", "line": 10, "start_line": 8, "body": "..." },
      { "in_reply_to": 123456789, "body": "reply text" }
    ]
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			var inputJSON string
			if jsonStr != "" {
				inputJSON = jsonStr
			} else {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				inputJSON = string(data)
			}
			if inputJSON == "" {
				return fmt.Errorf("no input: provide --json or pipe JSON to stdin")
			}

			var input reviewInput
			if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
			if len(input.Comments) == 0 && input.Event == "" {
				return fmt.Errorf("no-op: provide at least one comment or an event")
			}
			if input.Event != "" {
				switch strings.ToUpper(input.Event) {
				case "APPROVE", "REQUEST_CHANGES", "COMMENT":
					input.Event = strings.ToUpper(input.Event)
				default:
					return fmt.Errorf("invalid event %q: must be one of APPROVE, REQUEST_CHANGES, COMMENT", input.Event)
				}
			}

			out, err := executeReview(owner, repo, prNumber, input)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().StringVar(&jsonStr, "json", "", "Review JSON (event, body, comments[])")
	_ = cmd.MarkFlagRequired("pr")

	return cmd
}
