# review-comments

Manage inline pull request review comments — wraps `/repos/{owner}/{repo}/pulls/{pull_number}/comments` (per-comment operations) and `/repos/{owner}/{repo}/pulls/comments/{comment_id}`.

## Commands

| Command                                 | Description                                 | Required flags                  |
| --------------------------------------- | ------------------------------------------- | ------------------------------- |
| `gh aux review-comments list`           | List all inline review comments on a PR     | `--pr`                          |
| `gh aux review-comments get`            | Get a single inline review comment by ID    | `--id`                          |
| `gh aux review-comments update`         | Update the body of an inline review comment | `--id`, `--body`                |
| `gh aux review-comments add`            | Add an inline review comment to a PR        | `--pr`, `--body`                |
| `gh aux review-comments resolve-thread` | Resolve a pull request review thread        | `--thread-id` OR `--comment-id` |

## Flags

| Flag                   | Description                                                                                         | Default                        |
| ---------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------ |
| `--repo OWNER/REPO`    | Target repository                                                                                   | Current directory's git remote |
| `--pr <number>`        | Pull request number (`list`, `add`)                                                                 | —                              |
| `--id <id>`            | Inline review comment integer ID (`get`, `update`)                                                  | —                              |
| `--body "..."`         | Comment body text (`update`, `add`)                                                                 | —                              |
| `--in-reply-to <id>`   | Reply to an existing inline comment thread (`add` only); omit to post a standalone comment          | —                              |
| `--path "..."`         | File path for standalone comment (`add`, required when `--in-reply-to` is not set)                  | —                              |
| `--line <n>`           | Line number for standalone comment (`add`, required when `--in-reply-to` is not set)                | —                              |
| `--side LEFT\|RIGHT`   | Diff side — uppercase (`add`, required for standalone comment)                                      | —                              |
| `--thread-id <nodeId>` | Review thread GraphQL node ID (`resolve-thread`)                                                    | —                              |
| `--comment-id <id>`    | Review comment integer ID from GitHub URL `#discussion_r<id>` (`resolve-thread`)                    | —                              |

## Output

**`list`** → array of inline comment objects (empty array `[]` if none):

```json
[
  { "id": 123, "body": "...", "path": "src/foo.ts", "line": 42, "originalLine": 42, "author": { "login": "alice" }, "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "url": "https://github.com/..." }
]
```

**`get` / `update`** → single inline comment object (same shape as list element)

**`add`** → single inline comment object (same shape)

> **Note**: `add` calls `POST /pulls/{n}/comments` which internally creates a new review and immediately submits it (`state: COMMENTED`). Fails with **422** if a pending review already exists for the authenticated user.

**`resolve-thread`** → `{ "threadId": "RT_kwDO...", "isResolved": true }`

## Usage Patterns

**List all inline review comments on a PR:**

```sh
gh aux review-comments list --pr 123
```

**Get a single comment:**

```sh
gh aux review-comments get --id 987654321
```

**Update a comment body:**

```sh
gh aux review-comments update --id 987654321 --body "Updated: prefer early return here."
```

**Add a standalone inline comment:**

```sh
gh aux review-comments add --pr 123 --body "nit" --path src/foo.ts --line 42 --side RIGHT
```

**Reply to an existing thread:**

```sh
gh aux review-comments add --pr 123 --body "Addressed in the latest commit." --in-reply-to 987654321
```

**Resolve a thread by GraphQL node ID:**

```sh
gh aux review-comments resolve-thread --thread-id RT_kwDOA...
```

**Resolve a thread from a GitHub URL `#discussion_r<id>`:**

```sh
gh aux review-comments resolve-thread --comment-id 3146047744
```
