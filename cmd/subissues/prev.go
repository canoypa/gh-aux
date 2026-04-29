package subissues

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newPrevCmd() *cobra.Command {
	var subIssueNumber int

	cmd := &cobra.Command{
		Use:   "prev",
		Short: "Get the previous sibling sub-issue by position",
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

			targetPos := -1
			for _, s := range subIssues {
				if s.Number == subIssueNumber {
					targetPos = *s.Position
					break
				}
			}
			if targetPos < 0 {
				return fmt.Errorf("sub-issue #%d not found under issue #%d", subIssueNumber, issueNumber)
			}

			if targetPos == 0 {
				return writeJSON(os.Stdout, nil)
			}

			return writeJSON(os.Stdout, subIssues[targetPos-1])
		},
	}

	cmd.Flags().IntVar(&subIssueNumber, "sub-issue", 0, "Target sub-issue number")
	_ = cmd.MarkFlagRequired("sub-issue")

	return cmd
}
