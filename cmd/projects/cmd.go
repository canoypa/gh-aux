package projects

import (
	"github.com/spf13/cobra"
)

var repoFlag string

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage GitHub Projects (V2)",
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")

	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newUpdateFieldCmd())
	cmd.AddCommand(newClearFieldCmd())

	return cmd
}
