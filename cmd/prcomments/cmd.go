package prcomments

import (
	"github.com/spf13/cobra"
)

// repoFlag is the persistent --repo flag value shared across all subcommands.
var repoFlag string

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr-comments",
		Short: "Manage pull request comments",
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newReviewCmd())
	cmd.AddCommand(newReplyReviewCmd())
	cmd.AddCommand(newTimelineCmd())

	return cmd
}
