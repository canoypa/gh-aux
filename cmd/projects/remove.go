package projects

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	var projectNumber int
	var issueNumber int
	var prNumber int

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove an issue or pull request from a Project V2",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			if issueNumber == 0 && prNumber == 0 {
				return fmt.Errorf("one of --issue or --pr is required")
			}
			if issueNumber != 0 && prNumber != 0 {
				return fmt.Errorf("--issue and --pr are mutually exclusive")
			}

			restClient, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}
			graphqlClient, err := api.DefaultGraphQLClient()
			if err != nil {
				return fmt.Errorf("failed to create GraphQL client: %w", err)
			}

			contentNodeID, _, _, _, err := resolveContentNodeID(restClient, owner, repo, issueNumber, prNumber)
			if err != nil {
				return err
			}

			projectNodeID, err := resolveProjectNodeID(graphqlClient, owner, projectNumber)
			if err != nil {
				return err
			}

			itemID, err := resolveProjectItemID(graphqlClient, projectNodeID, contentNodeID)
			if err != nil {
				return err
			}

			// Fetch item details before deletion so we can return the full output.
			out, err := fetchProjectItem(graphqlClient, projectNodeID, itemID)
			if err != nil {
				return err
			}

			var result struct {
				DeleteProjectV2Item struct {
					DeletedItemID string `json:"deletedItemId"`
				} `json:"deleteProjectV2Item"`
			}
			mutation := `
mutation($projectId: ID!, $itemId: ID!) {
  deleteProjectV2Item(input: { projectId: $projectId, itemId: $itemId }) {
    deletedItemId
  }
}`
			if err := graphqlClient.Do(mutation, map[string]interface{}{
				"projectId": projectNodeID,
				"itemId":    itemID,
			}, &result); err != nil {
				return fmt.Errorf("failed to remove item from project: %w", err)
			}

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&projectNumber, "project", 0, "Project number")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "Issue number (mutually exclusive with --pr)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number (mutually exclusive with --issue)")
	_ = cmd.MarkFlagRequired("project")

	return cmd
}
