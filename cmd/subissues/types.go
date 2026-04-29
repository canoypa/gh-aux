package subissues

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// subIssueOutput is the normalized output for a sub-issue.
type subIssueOutput struct {
	Number   int    `json:"number"`
	Title    string `json:"title"`
	State    string `json:"state"`
	URL      string `json:"url"`
	Position *int   `json:"position,omitempty"`
}

// resolveRepo resolves owner and repository name from a "OWNER/REPO" string,
// falling back to the current directory's git remote if the argument is empty.
func resolveRepo(repoStr string) (owner, name string, err error) {
	if repoStr != "" {
		parts := strings.SplitN(repoStr, "/", 2)
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

// writeJSON serializes v as JSON to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// resolveIssueNodeID returns the GraphQL node ID of an issue via the REST API.
func resolveIssueNodeID(restClient *api.RESTClient, owner, repo string, issueNumber int) (string, error) {
	var result struct {
		NodeID string `json:"node_id"`
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, issueNumber)
	if err := restClient.Get(path, &result); err != nil {
		return "", fmt.Errorf("failed to get issue #%d: %w", issueNumber, err)
	}
	return result.NodeID, nil
}

// fetchSubIssues returns all sub-issues of a parent issue (by its node ID) in position order.
func fetchSubIssues(graphqlClient *api.GraphQLClient, parentNodeID string) ([]subIssueOutput, error) {
	type subIssueNode struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		URL    string `json:"url"`
	}

	var allNodes []subIssueNode
	var cursor *string

	for {
		var result struct {
			Node struct {
				SubIssues struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []subIssueNode `json:"nodes"`
				} `json:"subIssues"`
			} `json:"node"`
		}
		query := `
query($id: ID!, $cursor: String) {
  node(id: $id) {
    ... on Issue {
      subIssues(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          number
          title
          state
          url
        }
      }
    }
  }
}`
		if err := graphqlClient.Do(query, map[string]interface{}{"id": parentNodeID, "cursor": cursor}, &result); err != nil {
			return nil, fmt.Errorf("failed to fetch sub-issues: %w", err)
		}
		allNodes = append(allNodes, result.Node.SubIssues.Nodes...)
		if !result.Node.SubIssues.PageInfo.HasNextPage {
			break
		}
		c := result.Node.SubIssues.PageInfo.EndCursor
		cursor = &c
	}

	out := make([]subIssueOutput, len(allNodes))
	for i, n := range allNodes {
		pos := i
		out[i] = subIssueOutput{
			Number:   n.Number,
			Title:    n.Title,
			State:    n.State,
			URL:      n.URL,
			Position: &pos,
		}
	}
	return out, nil
}
