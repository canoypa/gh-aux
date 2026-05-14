package subissues

import (
	"fmt"

	"github.com/spf13/cobra"
)

var repoFlag string
var issueNumber int

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sub-issues",
		Short: "Manage sub-issues of a GitHub issue",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if issueNumber <= 0 {
				return fmt.Errorf("issue must be > 0")
			}
			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&repoFlag, "repo", "", "Repository in OWNER/REPO format (defaults to current directory remote)")
	cmd.PersistentFlags().IntVar(&issueNumber, "issue", 0, "Parent issue number")
	_ = cmd.MarkPersistentFlagRequired("issue")

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newPrevCmd())
	cmd.AddCommand(newNextCmd())
	cmd.AddCommand(newParentCmd())

	return cmd
}
