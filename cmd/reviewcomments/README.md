# review-comments

Manage inline pull request review comments.

Wraps `/repos/{owner}/{repo}/pulls/{pull_number}/comments` and `/repos/{owner}/{repo}/pulls/comments/{comment_id}`.

## Commands

- `list` — List all inline review comments on a PR
- `get` — Get a single inline review comment by ID
- `update` — Update the body of an inline review comment
- `add` — Add an inline review comment (standalone or reply)
- `resolve-thread` — Resolve a pull request review thread

See [skills reference](../../skills/gh-aux/references/review-comments.md) for full flag documentation.
