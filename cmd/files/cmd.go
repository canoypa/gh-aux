package files

import (
	"github.com/spf13/cobra"
)

// repoFlag is the persistent --repo flag value shared across all subcommands.
var repoFlag string

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "Get file contents from a remote repository",
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newDownloadCmd())

	return cmd
}
