package reviewcomments

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var prNumber int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inline review comments on a pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if prNumber <= 0 {
				return fmt.Errorf("pr must be > 0")
			}
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			var all []ReviewComment
			page := 1
			for {
				var raw []rawComment
				path := fmt.Sprintf("repos/%s/%s/pulls/%d/comments?per_page=100&page=%d", owner, repo, prNumber, page)
				if err := client.Get(path, &raw); err != nil {
					return fmt.Errorf("failed to list comments: %w", err)
				}
				for _, r := range raw {
					all = append(all, r.toReviewComment())
				}
				if len(raw) < 100 {
					break
				}
				page++
			}

			if all == nil {
				all = []ReviewComment{}
			}

			return writeJSON(os.Stdout, all)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	_ = cmd.MarkFlagRequired("pr")

	return cmd
}
