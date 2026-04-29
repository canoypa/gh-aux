package subissues

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newParentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parent",
		Short: "Get the parent issue of a sub-issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			restClient, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}
			graphqlClient, err := api.DefaultGraphQLClient()
			if err != nil {
				return fmt.Errorf("failed to create GraphQL client: %w", err)
			}

			nodeID, err := resolveIssueNodeID(restClient, owner, repo, issueNumber)
			if err != nil {
				return err
			}

			var result struct {
				Node struct {
					Parent *struct {
						Number int    `json:"number"`
						Title  string `json:"title"`
						State  string `json:"state"`
						URL    string `json:"url"`
					} `json:"parent"`
				} `json:"node"`
			}
			query := `
query($id: ID!) {
  node(id: $id) {
    ... on Issue {
      parent {
        number
        title
        state
        url
      }
    }
  }
}`
			if err := graphqlClient.Do(query, map[string]interface{}{"id": nodeID}, &result); err != nil {
				return fmt.Errorf("failed to get parent issue: %w", err)
			}

			if result.Node.Parent == nil {
				return writeJSON(os.Stdout, nil)
			}

			out := subIssueOutput{
				Number: result.Node.Parent.Number,
				Title:  result.Node.Parent.Title,
				State:  result.Node.Parent.State,
				URL:    result.Node.Parent.URL,
			}
			return writeJSON(os.Stdout, out)
		},
	}

	return cmd
}
