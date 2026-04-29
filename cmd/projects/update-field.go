package projects

import (
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newUpdateFieldCmd() *cobra.Command {
	var projectNumber int
	var issueNumber int
	var prNumber int
	var field string

	cmd := &cobra.Command{
		Use:   "update-field",
		Short: "Set a field value on a Project V2 item",
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

			eqIdx := strings.IndexByte(field, '=')
			if eqIdx < 0 {
				return fmt.Errorf("invalid --field value %q: expected FieldName=Value", field)
			}
			fieldName := field[:eqIdx]
			fieldValue := field[eqIdx+1:]

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

			if err := setFieldValue(graphqlClient, projectNodeID, itemID, fieldName, fieldValue); err != nil {
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
	cmd.Flags().StringVar(&field, "field", "", "Field value in FieldName=Value format")
	_ = cmd.MarkFlagRequired("project")
	_ = cmd.MarkFlagRequired("field")

	return cmd
}


// resolveProjectItemID finds the project item node ID for a given content node ID within a project.
func resolveProjectItemID(graphqlClient *api.GraphQLClient, projectNodeID, contentNodeID string) (string, error) {
	// Page through all items in the project looking for a matching content node ID.
	var cursor *string
	for {
		var result struct {
			Node struct {
				Items struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []struct {
						ID      string `json:"id"`
						Content struct {
							Typename string `json:"__typename"`
							ID       string `json:"id"`
						} `json:"content"`
					} `json:"nodes"`
				} `json:"items"`
			} `json:"node"`
		}
		query := `
query($projectId: ID!, $cursor: String) {
  node(id: $projectId) {
    ... on ProjectV2 {
      items(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          content {
            __typename
            ... on Issue       { id }
            ... on PullRequest { id }
          }
        }
      }
    }
  }
}`
		vars := map[string]interface{}{"projectId": projectNodeID, "cursor": cursor}
		if err := graphqlClient.Do(query, vars, &result); err != nil {
			return "", fmt.Errorf("failed to list project items: %w", err)
		}
		for _, node := range result.Node.Items.Nodes {
			if node.Content.ID == contentNodeID {
				return node.ID, nil
			}
		}
		if !result.Node.Items.PageInfo.HasNextPage {
			break
		}
		c := result.Node.Items.PageInfo.EndCursor
		cursor = &c
	}
	return "", fmt.Errorf("content not found in project")
}

// fetchProjectItem retrieves a project item and returns a projectItemOutput.
func fetchProjectItem(graphqlClient *api.GraphQLClient, projectID, itemID string) (projectItemOutput, error) {
	var result struct {
		Node struct {
			ID      string `json:"id"`
			Project struct {
				ID string `json:"id"`
			} `json:"project"`
			Content struct {
				Typename string `json:"__typename"`
				Number   int    `json:"number"`
				Title    string `json:"title"`
				URL      string `json:"url"`
			} `json:"content"`
		} `json:"node"`
	}
	query := `
query($itemId: ID!) {
  node(id: $itemId) {
    ... on ProjectV2Item {
      id
      project {
        id
      }
      content {
        __typename
        ... on Issue {
          number
          title
          url
        }
        ... on PullRequest {
          number
          title
          url
        }
      }
    }
  }
}`
	if err := graphqlClient.Do(query, map[string]interface{}{
		"itemId": itemID,
	}, &result); err != nil {
		return projectItemOutput{}, fmt.Errorf("failed to fetch project item: %w", err)
	}

	out := projectItemOutput{}
	out.ID = result.Node.ID
	out.ProjectID = result.Node.Project.ID
	out.Content.Number = result.Node.Content.Number
	out.Content.Title = result.Node.Content.Title
	out.Content.URL = result.Node.Content.URL
	return out, nil
}


