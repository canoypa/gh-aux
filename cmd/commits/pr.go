package commits

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

type prOutput struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	URL       string `json:"url"`
	MergedAt  string `json:"mergedAt"`
	Author    string `json:"author"`
}

func newPRCmd() *cobra.Command {
	var commitSHA string

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "List pull requests associated with a commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			restClient, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			var apiResponse []struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
				State  string `json:"state"`
				HTMLURL string `json:"html_url"`
				MergedAt string `json:"merged_at"`
				User struct {
					Login string `json:"login"`
				} `json:"user"`
			}

			err = restClient.Get(
				fmt.Sprintf("repos/%s/%s/commits/%s/pulls", owner, repo, commitSHA),
				&apiResponse,
			)
			if err != nil {
				return fmt.Errorf("failed to get pull requests for commit %s: %w", commitSHA, err)
			}

			out := make([]prOutput, len(apiResponse))
			for i, pr := range apiResponse {
				out[i] = prOutput{
					Number:   pr.Number,
					Title:    pr.Title,
					State:    pr.State,
					URL:      pr.HTMLURL,
					MergedAt: pr.MergedAt,
					Author:   pr.User.Login,
				}
			}

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().StringVar(&commitSHA, "commit", "", "Commit SHA")
	_ = cmd.MarkFlagRequired("commit")

	return cmd
}
