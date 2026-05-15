package reviews

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// reviewCommentInput is the input shape for a single review comment.
type reviewCommentInput struct {
	Path        string `json:"path,omitempty"`
	Line        int    `json:"line,omitempty"`
	StartLine   int    `json:"start_line,omitempty"`
	Side        string `json:"side,omitempty"`
	StartSide   string `json:"start_side,omitempty"`
	Body        string `json:"body"`
	InReplyTo   int    `json:"in_reply_to,omitempty"`
	SubjectType string `json:"subject_type,omitempty"`
}

// reviewInput is the input schema accepted by the create command.
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

// resolveRepo resolves owner and repository name from a "OWNER/REPO" string,
// falling back to the current directory's git remote if the argument is empty.
func resolveRepo(repoStr string) (owner, name string, err error) {
	if repoStr != "" {
		parts := strings.Split(repoStr, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", fmt.Errorf("invalid repo format %q: expected OWNER/REPO", repoStr)
		}
		return parts[0], parts[1], nil
	}
	r, err := repository.Current()
	if err != nil {
		return "", "", fmt.Errorf("could not determine current repository (use --repo): %w", err)
	}
	return r.Owner, r.Name, nil
}

// writeJSON writes v as indented JSON to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// validateComments checks reviewCommentInput fields before any API calls are made.
func validateComments(comments []reviewCommentInput) error {
	for i, c := range comments {
		if c.InReplyTo != 0 {
			if c.Body == "" {
				return fmt.Errorf("comments[%d]: body is required for replies", i)
			}
			continue
		}
		if c.Path == "" {
			return fmt.Errorf("comments[%d]: path is required", i)
		}
		if c.Body == "" {
			return fmt.Errorf("comments[%d]: body is required", i)
		}
		if c.SubjectType != "" && c.SubjectType != "file" {
			return fmt.Errorf("comments[%d]: subject_type must be \"file\" when set", i)
		}
		if c.SubjectType == "file" {
			continue
		}
		if c.Line <= 0 {
			return fmt.Errorf("comments[%d]: line must be > 0", i)
		}
		if c.Side != "LEFT" && c.Side != "RIGHT" {
			return fmt.Errorf("comments[%d]: side must be \"LEFT\" or \"RIGHT\" (uppercase)", i)
		}
		if c.StartLine != 0 && c.StartLine > c.Line {
			return fmt.Errorf("comments[%d]: start_line (%d) must be <= line (%d)", i, c.StartLine, c.Line)
		}
		if c.StartLine != 0 && c.StartSide == "" {
			return fmt.Errorf("comments[%d]: start_side is required when start_line is set", i)
		}
		if c.StartSide != "" && c.StartLine == 0 {
			return fmt.Errorf("comments[%d]: start_line is required when start_side is set", i)
		}
		if c.StartSide != "" && c.StartSide != "LEFT" && c.StartSide != "RIGHT" {
			return fmt.Errorf("comments[%d]: start_side must be \"LEFT\" or \"RIGHT\" (uppercase)", i)
		}
	}
	return nil
}

// findOrCreatePendingReview returns the current user's pending review,
// creating one (with the PR's head commit) if none exists.
// The returned bool is true when a new review was created (false = pre-existing).
func findOrCreatePendingReview(restClient *api.RESTClient, owner, repo string, prNumber int) (rawReviewResponse, bool, error) {
	var currentUser struct {
		Login string `json:"login"`
	}
	if err := restClient.Get("user", &currentUser); err != nil {
		return rawReviewResponse{}, false, fmt.Errorf("failed to get current user: %w", err)
	}

	reviewsPath := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", owner, repo, prNumber)

	for page := 1; ; page++ {
		var reviews []rawReviewResponse
		if err := restClient.Get(fmt.Sprintf("%s?per_page=100&page=%d", reviewsPath, page), &reviews); err != nil {
			return rawReviewResponse{}, false, fmt.Errorf("failed to get reviews: %w", err)
		}
		for _, r := range reviews {
			if r.State == "PENDING" && r.User.Login == currentUser.Login {
				return r, false, nil
			}
		}
		if len(reviews) < 100 {
			break
		}
	}

	var pr struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := restClient.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, prNumber), &pr); err != nil {
		return rawReviewResponse{}, false, fmt.Errorf("failed to get PR: %w", err)
	}

	reqBytes, err := json.Marshal(map[string]string{"commit_id": pr.Head.SHA})
	if err != nil {
		return rawReviewResponse{}, false, fmt.Errorf("failed to encode request: %w", err)
	}

	var created rawReviewResponse
	if err := restClient.Post(reviewsPath, bytes.NewReader(reqBytes), &created); err != nil {
		return rawReviewResponse{}, false, fmt.Errorf("failed to create pending review: %w", err)
	}

	return created, true, nil
}

// findThreadNodeID returns the GraphQL node_id of the review thread containing commentID.
func findThreadNodeID(gqlClient *api.GraphQLClient, owner, repo string, prNumber, commentID int) (string, error) {
	const query = `
query($owner: String!, $repo: String!, $pr: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
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
	type resultShape struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
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

	var after interface{} = nil
	for {
		var result resultShape
		if err := gqlClient.Do(query, map[string]interface{}{
			"owner": owner,
			"repo":  repo,
			"pr":    prNumber,
			"after": after,
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

		if !result.Repository.PullRequest.ReviewThreads.PageInfo.HasNextPage {
			break
		}
		after = result.Repository.PullRequest.ReviewThreads.PageInfo.EndCursor
	}

	return "", fmt.Errorf("thread containing comment %d not found", commentID)
}

// addCommentToReview adds a single inline comment to a pending review via GraphQL.
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

	if c.SubjectType == "file" {
		input := map[string]interface{}{
			"pullRequestReviewId": reviewNodeID,
			"path":                c.Path,
			"body":                c.Body,
			"subjectType":         "FILE",
		}
		const mutation = `
mutation($input: AddPullRequestReviewThreadInput!) {
  addPullRequestReviewThread(input: $input) {
    thread { id }
  }
}`
		var result struct {
			AddPullRequestReviewThread struct {
				Thread struct{ ID string `json:"id"` } `json:"thread"`
			} `json:"addPullRequestReviewThread"`
		}
		if err := gqlClient.Do(mutation, map[string]interface{}{"input": input}, &result); err != nil {
			return fmt.Errorf("failed to add file comment to %s: %w", c.Path, err)
		}
		return nil
	}

	input := map[string]interface{}{
		"pullRequestReviewId": reviewNodeID,
		"path":                c.Path,
		"line":                c.Line,
		"side":                c.Side,
		"body":                c.Body,
	}
	if c.StartLine != 0 {
		input["startLine"] = c.StartLine
		input["startSide"] = c.StartSide
	}

	const mutation = `
mutation($input: AddPullRequestReviewThreadInput!) {
  addPullRequestReviewThread(input: $input) {
    thread { id }
  }
}`
	var result struct {
		AddPullRequestReviewThread struct {
			Thread struct{ ID string `json:"id"` } `json:"thread"`
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

// executeReview is the core logic for the create command.
func executeReview(owner, repo string, prNumber int, input reviewInput) (reviewOutput, error) {
	restClient, err := api.DefaultRESTClient()
	if err != nil {
		return reviewOutput{}, fmt.Errorf("failed to create REST client: %w", err)
	}

	if input.Body != "" && input.Event == "" {
		return reviewOutput{}, fmt.Errorf("body requires event to be set")
	}
	if err := validateComments(input.Comments); err != nil {
		return reviewOutput{}, err
	}

	pending, created, err := findOrCreatePendingReview(restClient, owner, repo, prNumber)
	if err != nil {
		return reviewOutput{}, err
	}

	deletePending := func() {
		if !created {
			return
		}
		path := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews/%d", owner, repo, prNumber, pending.ID)
		_ = restClient.Delete(path, nil)
	}

	if len(input.Comments) > 0 {
		gqlClient, err := api.DefaultGraphQLClient()
		if err != nil {
			deletePending()
			return reviewOutput{}, fmt.Errorf("failed to create GraphQL client: %w", err)
		}
		for _, c := range input.Comments {
			if err := addCommentToReview(gqlClient, owner, repo, prNumber, pending.NodeID, c); err != nil {
				deletePending()
				return reviewOutput{}, err
			}
		}
	}

	if input.Event != "" {
		result, err := submitPendingReview(restClient, owner, repo, prNumber, pending.ID, input.Event, input.Body)
		if err != nil {
			deletePending()
			return reviewOutput{}, err
		}
		return result.toOutput(), nil
	}

	return pending.toOutput(), nil
}
