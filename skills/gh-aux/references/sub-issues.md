# sub-issues

List and manage sub-issues of a GitHub issue.

## Commands

| Command | Description | Required flags |
|---|---|---|
| `gh aux sub-issues list` | List all sub-issues of a parent issue | `--issue` |
| `gh aux sub-issues add` | Add a sub-issue to a parent issue | `--issue`, `--sub-issue` |
| `gh aux sub-issues remove` | Remove a sub-issue from a parent issue | `--issue`, `--sub-issue` |
| `gh aux sub-issues prev` | Get the previous sibling sub-issue by position | `--issue`, `--sub-issue` |
| `gh aux sub-issues next` | Get the next sibling sub-issue by position | `--issue`, `--sub-issue` |
| `gh aux sub-issues parent` | Get the parent issue of a sub-issue | `--issue` |

## Flags

| Flag | Description | Default |
|---|---|---|
| `--repo OWNER/REPO` | Target repository | Current directory's git remote |
| `--issue <number>` | For `list`/`add`/`remove`/`prev`/`next`: parent issue number. For `parent`: the child issue whose parent to retrieve. | — |
| `--sub-issue <number>` | Target sub-issue number | — |

## Output

`list` outputs a JSON array. All other commands output a single JSON object (or `null` when `prev`/`next` has no sibling):

```json
{
  "number": 12,
  "title": "Sub-task A",
  "state": "OPEN",
  "url": "https://github.com/owner/repo/issues/12",
  "position": 0
}
```

`state` is `"OPEN"` or `"CLOSED"`. `position` is 0-based. It is omitted from `add`/`remove` output. `parent` returns `null` when the issue has no parent.

## Usage patterns

```sh
# List all sub-issues of issue #5
gh aux sub-issues list --issue 5

# Add issue #12 as a sub-issue of #5
gh aux sub-issues add --issue 5 --sub-issue 12

# Remove sub-issue #12 from #5
gh aux sub-issues remove --issue 5 --sub-issue 12

# Get the issue that comes after #12 in the sub-issue list
gh aux sub-issues next --issue 5 --sub-issue 12

# Get the issue that comes before #12
gh aux sub-issues prev --issue 5 --sub-issue 12

# Get the parent issue of #12
gh aux sub-issues parent --issue 12
```
