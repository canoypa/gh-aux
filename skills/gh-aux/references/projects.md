# projects

Add issues and pull requests to a GitHub Project V2, and set field values.

## Commands

| Command | Description | Required flags |
|---|---|---|
| `gh aux projects add`          | Add an issue or PR to a Project V2, optionally set fields    | `--project`, one of `--issue` / `--pr` |
| `gh aux projects remove`       | Remove an issue or PR from a Project V2                      | `--project`, one of `--issue` / `--pr` |
| `gh aux projects update-field` | Set a field value on a project item                          | `--project`, one of `--issue` / `--pr`, `--field` |
| `gh aux projects clear-field`  | Clear a field value on a project item                        | `--project`, one of `--issue` / `--pr`, `--field-name` |

## Flags

| Flag | Description | Default |
|---|---|---|
| `--repo OWNER/REPO` | Target repository | Current directory's git remote |
| `--project <number>` | Project number as shown in the project's URL. The owner is resolved from `--repo` (or the current git remote); organization projects are searched first, then personal projects. | — |
| `--issue <number>` | Issue number | — |
| `--pr <number>` | Pull request number | — |
| `--field "Name=Value"` | Field assignment for `add` and `update-field`. `Value` is the option **name** (text) for `SINGLE_SELECT` fields, or the iteration **node ID** (e.g. `PVTI_...`) for `ITERATION` fields. Other types accept the raw value string. Repeatable (`add` only). | — |
| `--field-name "Name"` | Field name for `clear-field`. Do not include `=Value` — only the field name. | — |

## Output

`add`, `update-field`, `clear-field`, `remove` all output the same JSON object:

```json
{
  "id": "PVTI_xxx",
  "projectId": "PVT_xxx",
  "content": {
    "number": 42,
    "title": "Fix bug",
    "url": "https://github.com/owner/repo/issues/42"
  }
}
```

## Errors

| Error message | Cause | Recommended action |
|---|---|---|
| `option "X" not found in field "F"; available options: A, B, C` | `SINGLE_SELECT` value didn't match any option name (case-insensitive) | Use one of the listed option names exactly |
| `field "F" not found in project` | Field name didn't match any field in the project | Verify the field name; use `gh project field-list --owner OWNER --number PROJECT` to list fields |
| `unsupported field type "T" for field "F"` | Field has a type not supported by this tool | Set the field value manually via the GitHub UI |
| `project #N not found for OWNER` | Project number doesn't exist under the owner | Check the project number in the project URL |
| `content not found in project` | Issue/PR is not yet a member of the project | Run `gh aux projects add` first |

## Usage patterns

```sh
# Add issue #10 to project #3
gh aux projects add --project 3 --issue 10

# Add and immediately set a field
gh aux projects add --project 3 --issue 10 --field "Status=In Progress"

# Remove issue from project
gh aux projects remove --project 3 --issue 10

# Set a field value
gh aux projects update-field --project 3 --issue 10 --field "Status=Done"

# Clear a field value
gh aux projects clear-field --project 3 --issue 10 --field-name "Status"
```
