# commits

Query commit metadata.

## Commands

| Command | Description | Required flags |
|---|---|---|
| `gh aux commits pr` | List pull requests associated with a commit | `--commit` |

## Flags

| Flag | Description | Default |
|---|---|---|
| `--repo OWNER/REPO` | Target repository | Current directory's git remote |
| `--commit <sha>` | Commit SHA | — |

## Output

`pr` outputs a JSON array:

```json
[
  {
    "number": 42,
    "title": "Fix: update dependency",
    "state": "closed",
    "url": "https://github.com/owner/repo/pull/42",
    "mergedAt": "2024-01-15T10:30:00Z",
    "author": "octocat"
  }
]
```

`state` is `"open"` or `"closed"`. `mergedAt` is an empty string when the PR is not merged.

## Usage patterns

```sh
# Find the PR that introduced a specific commit
gh aux commits pr --commit abc1234

# With explicit repo
gh aux commits pr --repo owner/repo --commit abc1234
```
