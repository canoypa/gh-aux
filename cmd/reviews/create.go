package reviews

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var prNumber int
	var jsonStr string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create or submit a pull request review",
		Long: `Create or submit a pull request review.

Accepts JSON via --json flag or stdin.

Schema:
  {
    "event":    "APPROVE | REQUEST_CHANGES | COMMENT",  // omit to leave as pending
    "body":     "overall review comment",
    "comments": [
      { "path": "src/foo.ts", "line": 42, "body": "...", "side": "RIGHT" },
      { "path": "src/bar.ts", "line": 10, "start_line": 8, "body": "...", "side": "RIGHT", "start_side": "RIGHT" },
      { "in_reply_to": 123456789, "body": "reply text" }
    ]
  }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if prNumber <= 0 {
				return fmt.Errorf("pr must be > 0")
			}
			owner, repo, err := resolveRepo(repoFlag)
			if err != nil {
				return err
			}

			var inputJSON string
			if jsonStr != "" {
				inputJSON = jsonStr
			} else {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read stdin: %w", err)
				}
				inputJSON = string(data)
			}
			if inputJSON == "" {
				return fmt.Errorf("no input: provide --json or pipe JSON to stdin")
			}

			var input reviewInput
			if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
			if len(input.Comments) == 0 && input.Event == "" {
				return fmt.Errorf("no-op: provide at least one comment or an event")
			}
			if input.Event != "" {
				switch strings.ToUpper(input.Event) {
				case "APPROVE", "REQUEST_CHANGES", "COMMENT":
					input.Event = strings.ToUpper(input.Event)
				default:
					return fmt.Errorf("invalid event %q: must be one of APPROVE, REQUEST_CHANGES, COMMENT", input.Event)
				}
			}

			out, err := executeReview(owner, repo, prNumber, input)
			if err != nil {
				return err
			}

			return writeJSON(os.Stdout, out)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "Pull request number")
	cmd.Flags().StringVar(&jsonStr, "json", "", "Review JSON (event, body, comments[])")
	_ = cmd.MarkFlagRequired("pr")

	return cmd
}
