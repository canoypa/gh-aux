# pr-comments

Manage pull request comments — inline review comments and general PR comments.

## Commands

| Command | Description | Required flags |
|---|---|---|
| `gh aux pr-comments timeline` | List PR comments and review threads in chronological order | `--pr` |
| `gh aux pr-comments get` | Get a single inline review comment by REST ID | `--id` |
| `gh aux pr-comments add` | Add a general (non-diff) PR comment | `--pr`, `--body` |
| `gh aux pr-comments review` | Create and/or submit a PR review (JSON input) | `--pr`, `--json` or stdin |
| `gh aux pr-comments reply-review` | Reply to an existing inline review comment thread | `--pr`, `--id`, `--body` |

## Flags

| Flag | Description | Default |
|---|---|---|
| `--repo OWNER/REPO` | Target repository | Current directory's git remote |
| `--pr <number>` | Pull request number | — |
| `--id <id>` | Inline review comment ID (diff thread only; **not** a general PR comment ID) | — |
| `--body "..."` | Comment body text | — |
| `--json '{...}'` | Review input JSON (`review` only) | — |
| `--cursor <cursor>` | Pagination cursor (`timeline` only) | — |
| `--unresolved` | Only return unresolved threads (`timeline` only) | false |

## Output

All commands output JSON to stdout.

- `timeline` → `{ pageInfo, nodes[] }` — each node is one of:
  - `{ type: "comment", id, body, author: { login }, createdAt, url }` (IssueComment)
  - `{ type: "review", id, state, body, author: { login }, submittedAt, url, threads[] }` (PullRequestReview)
  - `threads[]` items: `{ isResolved, comments[] }`
  - `comments[]` items within a thread: `{ id, path, line?, originalLine?, diffHunk, body, author: { login }, createdAt, url }`
  - Bodies truncated to 200 chars.
  - All inline comments are fully fetched regardless of count (automatic pagination).
- `get` → single inline comment object: `{ id, body, path, line, originalLine, author: { login }, createdAt, updatedAt, url }`
- `add` → single general comment object: `{ id, body, author: { login }, createdAt, updatedAt, url }`
- `review` → review object: `{ id, state, body, author: { login }, url, submittedAt }`. `state` is `PENDING` when no `event` was provided.
- `reply-review` → single inline review comment object: `{ id, body, path, line, originalLine, author: { login }, createdAt, updatedAt, url }`. The reply is posted immediately (not held in a pending review).

## Review Input Schema (`review` command)

```json
{
  "event": "APPROVE | REQUEST_CHANGES | COMMENT",
  "body": "top-level review comment",
  "comments": [
    {
      "path": "src/foo.ts",
      "line": 42,
      "body": "Consider extracting this.",
      "start_line": 40,
      "side": "RIGHT",
      "in_reply_to": 123456789
    }
  ]
}
```

`event` and `body` are optional. Omit `event` to leave the review in pending state. Set `in_reply_to` to reply within an existing thread instead of starting a new one.

## Usage Patterns

**Get full PR conversation with unresolved threads only:**

```sh
gh aux pr-comments timeline --pr 123 --unresolved
```

**Submit an approval:**

```sh
echo '{"event":"APPROVE","body":"LGTM"}' | gh aux pr-comments review --pr 123
```

**Request changes with an inline comment:**

```sh
gh aux pr-comments review --pr 123 --json '{"event":"REQUEST_CHANGES","comments":[{"path":"src/foo.ts","line":42,"body":"Consider extracting this."}]}'
```

**Reply to an existing review thread:**

```sh
gh aux pr-comments reply-review --pr 123 --id 987654321 --body "Addressed in the latest commit."
```

**Add a general PR comment:**

```sh
gh aux pr-comments add --pr 123 --body "Please update the PR description to include a migration guide."
```
