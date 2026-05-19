package projects

import (
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newClearFieldCmd() *cobra.Command {
	var projectNumber int
	var issueNumber int
	var prNumber int
	var fieldNames []string

	cmd := &cobra.Command{
		Use:   "clear-field",
		Short: "Clear a field value on a Project V2 item",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			if projectNumber <= 0 {
				return fmt.Errorf("project must be > 0")
			}
			if issueNumber == 0 && prNumber == 0 {
				return fmt.Errorf("one of --issue or --pr is required")
			}
			if issueNumber != 0 && prNumber != 0 {
				return fmt.Errorf("--issue and --pr are mutually exclusive")
			}
			if len(fieldNames) == 0 {
				return fmt.Errorf("at least one --field-name is required")
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

			// Phase 1: fetch field definitions once and resolve all field IDs.
			fieldDefs, err := fetchProjectFields(graphqlClient, projectNodeID)
			if err != nil {
				return err
			}
			type resolvedClear struct {
				name    string
				fieldID string
			}
			resolved := make([]resolvedClear, 0, len(fieldNames))
			for _, name := range fieldNames {
				var found string
				for _, f := range fieldDefs {
					if strings.EqualFold(f.Name, name) {
						found = f.ID
						break
					}
				}
				if found == "" {
					return fmt.Errorf("field %q not found in project", name)
				}
				resolved = append(resolved, resolvedClear{name, found})
			}

			// Phase 2: clear all fields.
			for _, r := range resolved {
				if err := clearFieldValue(graphqlClient, projectNodeID, itemID, r.fieldID); err != nil {
					return fmt.Errorf("failed to clear field %q: %w", r.name, err)
				}
			}

			out, err := fetchProjectItem(graphqlClient, projectNodeID, itemID)
			if err != nil {
				return err
			}
			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&projectNumber, "project", 0, "Project number")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "Issue number (mutually exclusive with --pr)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number (mutually exclusive with --issue)")
	cmd.Flags().StringArrayVar(&fieldNames, "field-name", nil, "Field name to clear (repeatable)")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("field-name")

	return cmd
}

// clearFieldValue calls clearProjectV2ItemFieldValue to remove a field value.
func clearFieldValue(graphqlClient *api.GraphQLClient, projectNodeID, itemID, fieldID string) error {
	var result struct {
		ClearProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"clearProjectV2ItemFieldValue"`
	}
	mutation := `
mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!) {
  clearProjectV2ItemFieldValue(input: {
    projectId: $projectId
    itemId: $itemId
    fieldId: $fieldId
  }) {
    projectV2Item { id }
  }
}`
	if err := graphqlClient.Do(mutation, map[string]interface{}{
		"projectId": projectNodeID,
		"itemId":    itemID,
		"fieldId":   fieldID,
	}, &result); err != nil {
		return fmt.Errorf("failed to clear field value: %w", err)
	}
	return nil
}
