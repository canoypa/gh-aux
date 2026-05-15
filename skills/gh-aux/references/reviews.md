# reviews

Create and submit pull request reviews — wraps `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews`.

## Commands

| Command                 | Description                            | Required flags            |
| ----------------------- | -------------------------------------- | ------------------------- |
| `gh aux reviews create` | Create or submit a pull request review | `--pr`, `--json` or stdin |

## Flags

| Flag                | Description                    | Default                        |
| ------------------- | ------------------------------ | ------------------------------ |
| `--repo OWNER/REPO` | Target repository              | Current directory's git remote |
| `--pr <number>`     | Pull request number            | —                              |
| `--json '{...}'`    | Review JSON (see schema below) | —                              |

## Output

```json
{ "id": 1234567890, "state": "APPROVED", "body": "LGTM", "author": { "login": "alice" }, "url": "https://github.com/...", "submittedAt": "2024-01-01T00:00:00Z" }
```

`state` is `"PENDING"` when no `event` was provided.

## Input Schema

```json
{
  "event":    "APPROVE | REQUEST_CHANGES | COMMENT",
  "body":     "top-level review comment",
  "comments": [
    {
      "path": "src/foo.ts",
      "line": 42,
      "body": "Consider extracting this.",
      "start_line": 40,
      "side": "RIGHT",
      "start_side": "RIGHT",
      "in_reply_to": 123456789,
      "subject_type": "file"
    }
  ]
}
```

- `event` — optional. Omit to leave review in `PENDING` state. If a pending review already exists, inline comments are appended to it; otherwise a new pending review is created.
- `body` — requires `event`. Top-level review comment submitted with the event.
- `in_reply_to` — set to reply within an existing thread instead of starting a new one.
- `subject_type: "file"` — post a file-level comment; no `line` or `side` needed.
- `side`, `start_side` — must be uppercase (`"LEFT"` or `"RIGHT"`).

## Usage Patterns

**Submit an approval:**

```sh
echo '{"event":"APPROVE","body":"LGTM"}' | gh aux reviews create --pr 123
```

**Request changes with an inline comment:**

```sh
gh aux reviews create --pr 123 --json '{"event":"REQUEST_CHANGES","comments":[{"path":"src/foo.ts","line":42,"body":"Consider extracting this.","side":"RIGHT"}]}'
```

**Add inline comments without submitting (pending):**

```sh
gh aux reviews create --pr 123 --json '{"comments":[{"path":"src/foo.ts","line":42,"body":"nit","side":"RIGHT"}]}'
```

**Reply within an existing thread (pending):**

```sh
gh aux reviews create --pr 123 --json '{"comments":[{"body":"Addressed in latest commit.","in_reply_to":987654321}]}'
```
