package projects

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	var projectNumber int
	var issueNumber int
	var prNumber int
	var fields []string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an issue or pull request to a Project V2",
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

			// Resolve the content node ID via REST.
			contentNodeID, contentNumber, contentTitle, contentURL, err := resolveContentNodeID(restClient, owner, repo, issueNumber, prNumber)
			if err != nil {
				return err
			}

			// Resolve the project node ID via GraphQL.
			projectNodeID, err := resolveProjectNodeID(graphqlClient, owner, projectNumber)
			if err != nil {
				return err
			}

			// Add item to project.
			itemNodeID, err := addItemToProject(graphqlClient, projectNodeID, contentNodeID)
			if err != nil {
				return err
			}

			// Set fields if specified. On failure, remove the item to avoid leaving it in a partial state.
			for _, f := range fields {
				eqIdx := strings.IndexByte(f, '=')
				if eqIdx < 0 {
					_ = removeItemFromProject(graphqlClient, projectNodeID, itemNodeID)
					return fmt.Errorf("invalid --field value %q: expected FieldName=Value", f)
				}
				fieldName := f[:eqIdx]
				fieldValue := f[eqIdx+1:]
				if err := setFieldValue(graphqlClient, projectNodeID, itemNodeID, fieldName, fieldValue); err != nil {
					_ = removeItemFromProject(graphqlClient, projectNodeID, itemNodeID)
					return fmt.Errorf("failed to set field %q: %w", fieldName, err)
				}
			}

			out := projectItemOutput{}
			out.ID = itemNodeID
			out.ProjectID = projectNodeID
			out.Content.Number = contentNumber
			out.Content.Title = contentTitle
			out.Content.URL = contentURL

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&projectNumber, "project", 0, "Project number")
	cmd.Flags().IntVar(&issueNumber, "issue", 0, "Issue number (mutually exclusive with --pr)")
	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number (mutually exclusive with --issue)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Field value in FieldName=Value format (repeatable)")
	_ = cmd.MarkFlagRequired("project")

	return cmd
}

// resolveContentNodeID returns the GraphQL node ID, number, title, and URL of an issue or PR.
func resolveContentNodeID(restClient *api.RESTClient, owner, repoName string, issueNumber, prNumber int) (nodeID string, number int, title string, url string, err error) {
	// Both issues and PRs are accessible via the issues endpoint for node_id.
	n := issueNumber
	if n == 0 {
		n = prNumber
	}

	var result struct {
		NodeID  string `json:"node_id"`
		Number  int    `json:"number"`
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
	}
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repoName, n)
	if err = restClient.Get(path, &result); err != nil {
		err = fmt.Errorf("failed to get issue/PR %d: %w", n, err)
		return
	}
	nodeID = result.NodeID
	number = result.Number
	title = result.Title
	url = result.HTMLURL
	return
}

// resolveProjectNodeID returns the GraphQL node ID of a Project V2 by number,
// trying organization first then user.
func resolveProjectNodeID(graphqlClient *api.GraphQLClient, owner string, projectNumber int) (string, error) {
	// Try organization project first.
	var orgResult struct {
		Organization struct {
			ProjectV2 struct {
				ID string `json:"id"`
			} `json:"projectV2"`
		} `json:"organization"`
	}
	orgQuery := `
query($owner: String!, $number: Int!) {
  organization(login: $owner) {
    projectV2(number: $number) {
      id
    }
  }
}`
	orgErr := graphqlClient.Do(orgQuery, map[string]interface{}{
		"owner":  owner,
		"number": projectNumber,
	}, &orgResult)
	if orgErr == nil && orgResult.Organization.ProjectV2.ID != "" {
		return orgResult.Organization.ProjectV2.ID, nil
	}

	// Fall back to user project.
	var userResult struct {
		User struct {
			ProjectV2 struct {
				ID string `json:"id"`
			} `json:"projectV2"`
		} `json:"user"`
	}
	userQuery := `
query($owner: String!, $number: Int!) {
  user(login: $owner) {
    projectV2(number: $number) {
      id
    }
  }
}`
	if err := graphqlClient.Do(userQuery, map[string]interface{}{
		"owner":  owner,
		"number": projectNumber,
	}, &userResult); err != nil {
		return "", fmt.Errorf("failed to resolve project #%d for %s: %w", projectNumber, owner, err)
	}
	if userResult.User.ProjectV2.ID == "" {
		return "", fmt.Errorf("project #%d not found for %s", projectNumber, owner)
	}
	return userResult.User.ProjectV2.ID, nil
}

// addItemToProject calls addProjectV2ItemById and returns the new item's node ID.
func addItemToProject(graphqlClient *api.GraphQLClient, projectID, contentID string) (string, error) {
	var result struct {
		AddProjectV2ItemByID struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}
	mutation := `
mutation($projectId: ID!, $contentId: ID!) {
  addProjectV2ItemById(input: { projectId: $projectId, contentId: $contentId }) {
    item {
      id
    }
  }
}`
	if err := graphqlClient.Do(mutation, map[string]interface{}{
		"projectId": projectID,
		"contentId": contentID,
	}, &result); err != nil {
		return "", fmt.Errorf("failed to add item to project: %w", err)
	}
	return result.AddProjectV2ItemByID.Item.ID, nil
}

// removeItemFromProject calls deleteProjectV2Item. Used for rollback when field-setting fails after add.
func removeItemFromProject(graphqlClient *api.GraphQLClient, projectID, itemID string) error {
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
	return graphqlClient.Do(mutation, map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
	}, &result)
}

// setFieldValue sets a named field on a project item.
// It resolves field ID and type via a GraphQL query, then calls updateProjectV2ItemFieldValue.
func setFieldValue(graphqlClient *api.GraphQLClient, projectID, itemID, fieldName, value string) error {
	// Fetch project fields.
	var fieldsResult struct {
		Node struct {
			Fields struct {
				Nodes []struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					DataType string `json:"dataType"`
					Options  []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"options"`
				} `json:"nodes"`
			} `json:"fields"`
		} `json:"node"`
	}
	fieldsQuery := `
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 100) {
        nodes {
          ... on ProjectV2Field {
            id
            name
            dataType
          }
          ... on ProjectV2SingleSelectField {
            id
            name
            dataType
            options {
              id
              name
            }
          }
          ... on ProjectV2IterationField {
            id
            name
            dataType
          }
        }
      }
    }
  }
}`
	if err := graphqlClient.Do(fieldsQuery, map[string]interface{}{
		"projectId": projectID,
	}, &fieldsResult); err != nil {
		return fmt.Errorf("failed to fetch project fields: %w", err)
	}

	// Find the matching field by name (case-insensitive).
	var fieldID string
	var dataType string
	var options []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	for _, f := range fieldsResult.Node.Fields.Nodes {
		if strings.EqualFold(f.Name, fieldName) {
			fieldID = f.ID
			dataType = f.DataType
			options = f.Options
			break
		}
	}
	if fieldID == "" {
		return fmt.Errorf("field %q not found in project", fieldName)
	}

	// Build the value union based on dataType.
	var fieldValueInput map[string]interface{}
	switch strings.ToUpper(dataType) {
	case "SINGLE_SELECT":
		optionID := ""
		for _, o := range options {
			if strings.EqualFold(o.Name, value) {
				optionID = o.ID
				break
			}
		}
		if optionID == "" {
			names := make([]string, 0, len(options))
			for _, o := range options {
				names = append(names, o.Name)
			}
			return fmt.Errorf("option %q not found in field %q; available options: %s", value, fieldName, strings.Join(names, ", "))
		}
		fieldValueInput = map[string]interface{}{"singleSelectOptionId": optionID}
	case "TEXT":
		fieldValueInput = map[string]interface{}{"text": value}
	case "NUMBER":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid number value %q for field %q: %w", value, fieldName, err)
		}
		fieldValueInput = map[string]interface{}{"number": n}
	case "DATE":
		fieldValueInput = map[string]interface{}{"date": value}
	case "ITERATION":
		fieldValueInput = map[string]interface{}{"iterationId": value}
	default:
		return fmt.Errorf("unsupported field type %q for field %q", dataType, fieldName)
	}

	var updateResult struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"updateProjectV2ItemFieldValue"`
	}
	updateMutation := `
mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: ProjectV2FieldValue!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $projectId
    itemId: $itemId
    fieldId: $fieldId
    value: $value
  }) {
    projectV2Item {
      id
    }
  }
}`
	if err := graphqlClient.Do(updateMutation, map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"value":     fieldValueInput,
	}, &updateResult); err != nil {
		return fmt.Errorf("failed to update field value: %w", err)
	}
	return nil
}
