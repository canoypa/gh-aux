package reviewcomments

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var commentID int

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a single inline review comment by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commentID <= 0 {
				return fmt.Errorf("id must be > 0")
			}
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			client, err := api.DefaultRESTClient()
			if err != nil {
				return fmt.Errorf("failed to create REST client: %w", err)
			}

			var raw rawComment
			path := fmt.Sprintf("repos/%s/%s/pulls/comments/%d", owner, repo, commentID)
			if err := client.Get(path, &raw); err != nil {
				return fmt.Errorf("failed to get comment %d: %w", commentID, err)
			}

			return writeJSON(os.Stdout, raw.toReviewComment())
		},
	}

	cmd.Flags().IntVar(&commentID, "id", 0, "Inline review comment ID (from a diff thread)")
	_ = cmd.MarkFlagRequired("id")

	return cmd
}
