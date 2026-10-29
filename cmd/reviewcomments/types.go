package reviewcomments

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// ReviewComment represents a pull request review comment (normalized to camelCase).
type ReviewComment struct {
	ID           int    `json:"id"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	Line         int    `json:"line"`
	OriginalLine int    `json:"originalLine"`
	Author       struct {
		Login string `json:"login"`
	} `json:"author"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	URL       string `json:"url"`
}

// rawComment is the REST API response shape for a pull request review comment.
type rawComment struct {
	ID           int    `json:"id"`
	Body         string `json:"body"`
	Path         string `json:"path"`
	Line         *int   `json:"line"`
	OriginalLine *int   `json:"original_line"`
	User         struct {
		Login string `json:"login"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	HTMLURL   string `json:"html_url"`
}

// toReviewComment converts a rawComment to a ReviewComment.
func (r rawComment) toReviewComment() ReviewComment {
	c := ReviewComment{
		ID:        r.ID,
		Body:      r.Body,
		Path:      r.Path,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		URL:       r.HTMLURL,
	}
	c.Author.Login = r.User.Login
	if r.Line != nil {
		c.Line = *r.Line
	}
	if r.OriginalLine != nil {
		c.OriginalLine = *r.OriginalLine
	}
	return c
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

// parsePullRequestNumber extracts the pull request number from the
// pull_request_url of a REST review comment
// (e.g. "https://api.github.com/repos/OWNER/REPO/pulls/1").
func parsePullRequestNumber(pullRequestURL string) (int, error) {
	i := strings.LastIndex(pullRequestURL, "/pulls/")
	if i < 0 {
		return 0, fmt.Errorf("invalid pull_request_url %q", pullRequestURL)
	}
	n, err := strconv.Atoi(pullRequestURL[i+len("/pulls/"):])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid pull_request_url %q", pullRequestURL)
	}
	return n, nil
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type threadComments struct {
	PageInfo pageInfo `json:"pageInfo"`
	Nodes    []struct {
		DatabaseID int `json:"databaseId"`
	} `json:"nodes"`
}

func (c threadComments) contains(commentID int) bool {
	for _, n := range c.Nodes {
		if n.DatabaseID == commentID {
			return true
		}
	}
	return false
}

// findThreadNodeID returns the GraphQL node ID of the review thread containing commentID.
func findThreadNodeID(client *api.GraphQLClient, owner, repo string, prNumber, commentID int) (string, error) {
	const query = `
query($owner: String!, $repo: String!, $pr: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $pr) {
      reviewThreads(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          comments(first: 100) {
            pageInfo { hasNextPage endCursor }
            nodes { databaseId }
          }
        }
      }
    }
  }
}`
	var after interface{} = nil
	for {
		var result struct {
			Repository struct {
				PullRequest struct {
					ReviewThreads struct {
						PageInfo pageInfo `json:"pageInfo"`
						Nodes    []struct {
							ID       string         `json:"id"`
							Comments threadComments `json:"comments"`
						} `json:"nodes"`
					} `json:"reviewThreads"`
				} `json:"pullRequest"`
			} `json:"repository"`
		}
		if err := client.Do(query, map[string]interface{}{
			"owner": owner,
			"repo":  repo,
			"pr":    prNumber,
			"after": after,
		}, &result); err != nil {
			return "", fmt.Errorf("failed to query review threads: %w", err)
		}

		threads := result.Repository.PullRequest.ReviewThreads
		for _, thread := range threads.Nodes {
			found, err := threadContainsComment(client, thread.ID, thread.Comments, commentID)
			if err != nil {
				return "", err
			}
			if found {
				return thread.ID, nil
			}
		}

		if !threads.PageInfo.HasNextPage {
			break
		}
		after = threads.PageInfo.EndCursor
	}

	return "", fmt.Errorf("review thread containing comment %d not found", commentID)
}

// threadContainsComment reports whether the thread contains commentID,
// fetching the comments that did not fit in the first page.
func threadContainsComment(client *api.GraphQLClient, threadID string, first threadComments, commentID int) (bool, error) {
	const query = `
query($id: ID!, $after: String) {
  node(id: $id) {
    ... on PullRequestReviewThread {
      comments(first: 100, after: $after) {
        pageInfo { hasNextPage endCursor }
        nodes { databaseId }
      }
    }
  }
}`
	comments := first
	for {
		if comments.contains(commentID) {
			return true, nil
		}
		if !comments.PageInfo.HasNextPage {
			return false, nil
		}
		var result struct {
			Node struct {
				Comments threadComments `json:"comments"`
			} `json:"node"`
		}
		if err := client.Do(query, map[string]interface{}{
			"id":    threadID,
			"after": comments.PageInfo.EndCursor,
		}, &result); err != nil {
			return false, fmt.Errorf("failed to query comments of review thread %s: %w", threadID, err)
		}
		comments = result.Node.Comments
	}
}
