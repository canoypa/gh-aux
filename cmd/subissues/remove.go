package subissues

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var subIssueNumber int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a sub-issue from a parent issue",
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
				RemoveSubIssue struct {
					SubIssue struct {
						Number int    `json:"number"`
						Title  string `json:"title"`
						State  string `json:"state"`
						URL    string `json:"url"`
					} `json:"subIssue"`
				} `json:"removeSubIssue"`
			}
			mutation := `
mutation($issueId: ID!, $subIssueId: ID!) {
  removeSubIssue(input: { issueId: $issueId, subIssueId: $subIssueId }) {
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
				return fmt.Errorf("failed to remove sub-issue: %w", err)
			}

			out := subIssueOutput{
				Number: result.RemoveSubIssue.SubIssue.Number,
				Title:  result.RemoveSubIssue.SubIssue.Title,
				State:  result.RemoveSubIssue.SubIssue.State,
				URL:    result.RemoveSubIssue.SubIssue.URL,
			}
			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&subIssueNumber, "sub-issue", 0, "Sub-issue number to remove")
	_ = cmd.MarkFlagRequired("sub-issue")

	return cmd
}
