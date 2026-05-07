package commits

import (
	"github.com/spf13/cobra"
)

var repoFlag string

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commits",
		Short: "Query commit metadata",
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")

	cmd.AddCommand(newPRCmd())

	return cmd
}
