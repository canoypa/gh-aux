# pr-comments

Manage pull request comments — inline review comments and general PR comments.

## Commands

### `pr-comments timeline`

List PR issue comments and review threads in chronological order.

```sh
gh aux pr-comments timeline --pr <number> [--cursor <cursor>] [--unresolved] [--repo OWNER/REPO]
```

### `pr-comments get`

Get a single inline review comment (diff thread comment) by its REST ID. **Not** usable for general PR comments created by `pr-comments add`.

```sh
gh aux pr-comments get --id <id> [--repo OWNER/REPO]
```

### `pr-comments add`

Add a general (non-diff) comment to a pull request.

```sh
gh aux pr-comments add --pr <number> --body "..." [--repo OWNER/REPO]
```

### `pr-comments review`

Create and/or submit a pull request review. Accepts JSON input via `--json` or stdin.
Input schema: `{ event?, body?, comments[] }` where each comment is `{ path, line, body, in_reply_to?, start_line?, side? }`.

If a pending review already exists it is reused; otherwise a new pending review is created.
Omit `event` to leave the review in pending state.

```sh
gh aux pr-comments review --pr <number> [--json '{...}'] [--repo OWNER/REPO]
# or pipe JSON via stdin
echo '{"event":"COMMENT","body":"LGTM"}' | gh aux pr-comments review --pr <number>
```

### `pr-comments reply-review`

Reply to an existing inline review comment thread. The reply is posted immediately and returns the created comment object.

```sh
gh aux pr-comments reply-review --pr <number> --id <comment-id> --body "..." [--repo OWNER/REPO]
```

