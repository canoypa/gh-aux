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

			if projectNumber <= 0 {
				return fmt.Errorf("project must be > 0")
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

// Pre-validate all field formats before making any API calls.
		for _, f := range fields {
			if strings.IndexByte(f, '=') < 0 {
				return fmt.Errorf("invalid --field value %q: expected FieldName=Value", f)
			}
		}

		// Set fields if specified. On failure, remove the item to avoid leaving it in a partial state.
		for _, f := range fields {
			eqIdx := strings.IndexByte(f, '=')
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

// projectField holds the definition of a single field in a ProjectV2.
type projectField struct {
	ID       string
	Name     string
	DataType string
	Options  []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
}

// fetchProjectFields returns all fields defined in a ProjectV2 (up to 100).
func fetchProjectFields(graphqlClient *api.GraphQLClient, projectID string) ([]projectField, error) {
	var result struct {
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
	query := `
query($projectId: ID!) {
  node(id: $projectId) {
    ... on ProjectV2 {
      fields(first: 100) {
        nodes {
          ... on ProjectV2Field {
            id name dataType
          }
          ... on ProjectV2SingleSelectField {
            id name dataType
            options { id name }
          }
          ... on ProjectV2IterationField {
            id name dataType
          }
        }
      }
    }
  }
}`
	if err := graphqlClient.Do(query, map[string]interface{}{"projectId": projectID}, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch project fields: %w", err)
	}
	fields := make([]projectField, 0, len(result.Node.Fields.Nodes))
	for _, n := range result.Node.Fields.Nodes {
		fields = append(fields, projectField{
			ID:       n.ID,
			Name:     n.Name,
			DataType: n.DataType,
			Options:  n.Options,
		})
	}
	return fields, nil
}

// resolveFieldValueInput finds the matching field and builds the value union without making any API calls.
func resolveFieldValueInput(fields []projectField, fieldName, value string) (fieldID string, valueUnion map[string]interface{}, err error) {
	var matched *projectField
	for i := range fields {
		if strings.EqualFold(fields[i].Name, fieldName) {
			matched = &fields[i]
			break
		}
	}
	if matched == nil {
		return "", nil, fmt.Errorf("field %q not found in project", fieldName)
	}
	switch strings.ToUpper(matched.DataType) {
	case "SINGLE_SELECT":
		optionID := ""
		for _, o := range matched.Options {
			if strings.EqualFold(o.Name, value) {
				optionID = o.ID
				break
			}
		}
		if optionID == "" {
			names := make([]string, 0, len(matched.Options))
			for _, o := range matched.Options {
				names = append(names, o.Name)
			}
			return "", nil, fmt.Errorf("option %q not found in field %q; available options: %s", value, fieldName, strings.Join(names, ", "))
		}
		return matched.ID, map[string]interface{}{"singleSelectOptionId": optionID}, nil
	case "TEXT":
		return matched.ID, map[string]interface{}{"text": value}, nil
	case "NUMBER":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", nil, fmt.Errorf("invalid number value %q for field %q: %w", value, fieldName, err)
		}
		return matched.ID, map[string]interface{}{"number": n}, nil
	case "DATE":
		return matched.ID, map[string]interface{}{"date": value}, nil
	case "ITERATION":
		return matched.ID, map[string]interface{}{"iterationId": value}, nil
	default:
		return "", nil, fmt.Errorf("unsupported field type %q for field %q", matched.DataType, fieldName)
	}
}

// applyFieldValue writes a resolved field value to a project item.
func applyFieldValue(graphqlClient *api.GraphQLClient, projectID, itemID, fieldID string, valueUnion map[string]interface{}) error {
	var updateResult struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"updateProjectV2ItemFieldValue"`
	}
	mutation := `
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
	if err := graphqlClient.Do(mutation, map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"value":     valueUnion,
	}, &updateResult); err != nil {
		return fmt.Errorf("failed to update field value: %w", err)
	}
	return nil
}

// setFieldValue sets a named field on a project item.
func setFieldValue(graphqlClient *api.GraphQLClient, projectID, itemID, fieldName, value string) error {
	fields, err := fetchProjectFields(graphqlClient, projectID)
	if err != nil {
		return err
	}
	fieldID, valueUnion, err := resolveFieldValueInput(fields, fieldName, value)
	if err != nil {
		return err
	}
	return applyFieldValue(graphqlClient, projectID, itemID, fieldID, valueUnion)
}
