package subissues

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var subIssueNumber int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a sub-issue to a parent issue",
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

			parentNodeID, err := resolveIssueNodeID(restClient, owner, repo, issueNumber)
			if err != nil {
				return err
			}
			childNodeID, err := resolveIssueNodeID(restClient, owner, repo, subIssueNumber)
			if err != nil {
				return err
			}

			var result struct {
				AddSubIssue struct {
					SubIssue struct {
						Number int    `json:"number"`
						Title  string `json:"title"`
						State  string `json:"state"`
						URL    string `json:"url"`
					} `json:"subIssue"`
				} `json:"addSubIssue"`
			}
			mutation := `
mutation($issueId: ID!, $subIssueId: ID!) {
  addSubIssue(input: { issueId: $issueId, subIssueId: $subIssueId }) {
    subIssue {
      number
      title
      state
      url
    }
  }
}`
			if err := graphqlClient.Do(mutation, map[string]interface{}{
				"issueId":    parentNodeID,
				"subIssueId": childNodeID,
			}, &result); err != nil {
				return fmt.Errorf("failed to add sub-issue: %w", err)
			}

			out := subIssueOutput{
				Number: result.AddSubIssue.SubIssue.Number,
				Title:  result.AddSubIssue.SubIssue.Title,
				State:  result.AddSubIssue.SubIssue.State,
				URL:    result.AddSubIssue.SubIssue.URL,
			}
			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&subIssueNumber, "sub-issue", 0, "Sub-issue number to add")
	_ = cmd.MarkFlagRequired("sub-issue")

	return cmd
}
