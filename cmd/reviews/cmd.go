package reviews

import (
	"github.com/spf13/cobra"
)

// repoFlag is the persistent --repo flag value shared across all subcommands.
var repoFlag string

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "Manage pull request reviews",
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")

	cmd.AddCommand(newCreateCmd())

	return cmd
}
