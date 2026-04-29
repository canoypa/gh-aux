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
	var fieldName string

	cmd := &cobra.Command{
		Use:   "clear-field",
		Short: "Clear a field value on a Project V2 item",
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

			fieldID, err := resolveFieldID(graphqlClient, projectNodeID, fieldName)
			if err != nil {
				return err
			}

			if err := clearFieldValue(graphqlClient, projectNodeID, itemID, fieldID); err != nil {
				return err
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
	cmd.Flags().StringVar(&fieldName, "field-name", "", "Field name to clear")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("field-name")

	return cmd
}

// resolveFieldID returns the node ID of a named field in a project (case-insensitive).
func resolveFieldID(graphqlClient *api.GraphQLClient, projectNodeID, fieldName string) (string, error) {
	var result struct {
		Node struct {
			Fields struct {
				Nodes []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"fields"`
		} `json:"node"`
	}
	query := `
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 100) {
        nodes {
          ... on ProjectV2Field             { id name }
          ... on ProjectV2SingleSelectField { id name }
          ... on ProjectV2IterationField    { id name }
        }
      }
    }
  }
}`
	if err := graphqlClient.Do(query, map[string]interface{}{"projectId": projectNodeID}, &result); err != nil {
		return "", fmt.Errorf("failed to fetch project fields: %w", err)
	}
	for _, f := range result.Node.Fields.Nodes {
		if strings.EqualFold(f.Name, fieldName) {
			return f.ID, nil
		}
	}
	return "", fmt.Errorf("field %q not found in project", fieldName)
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
