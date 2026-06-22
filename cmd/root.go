package cmd

import (
	"strings"

	"github.com/canoypa/gh-aux/cmd/commits"
	"github.com/canoypa/gh-aux/cmd/projects"
	"github.com/canoypa/gh-aux/cmd/reviewcomments"
	"github.com/canoypa/gh-aux/cmd/reviews"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aux",
	Short: "gh auxiliary commands",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(commits.NewCmd())
	rootCmd.AddCommand(projects.NewCmd())
	rootCmd.AddCommand(reviewcomments.NewCmd())
	rootCmd.AddCommand(reviews.NewCmd())

	// Patch usage template so all commands display as "gh aux ..." instead of "aux ...".
	t := rootCmd.UsageTemplate()
	t = strings.ReplaceAll(t, "{{.UseLine}}", "gh {{.UseLine}}")
	t = strings.ReplaceAll(t, "{{.CommandPath}} [command]", "gh {{.CommandPath}} [command]")
	t = strings.ReplaceAll(t, `"{{.CommandPath}} [command] --help"`, `"gh {{.CommandPath}} [command] --help"`)
	rootCmd.SetUsageTemplate(t)
}
