package subissues

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sub-issues of a parent issue",
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

			subIssues, err := fetchSubIssues(graphqlClient, parentNodeID)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, subIssues)
		},
	}
	return cmd
}
