# sub-issues

List and manage sub-issues of a GitHub issue.

## Commands

### `sub-issues list`

List all sub-issues of a parent issue, with position index.

```sh
gh aux sub-issues list --issue <number> [--repo OWNER/REPO]
```

### `sub-issues add`

Add a sub-issue to a parent issue.

```sh
gh aux sub-issues add --issue <parent-number> --sub-issue <child-number> [--repo OWNER/REPO]
```

### `sub-issues remove`

Remove a sub-issue from a parent issue.

```sh
gh aux sub-issues remove --issue <parent-number> --sub-issue <child-number> [--repo OWNER/REPO]
```

### `sub-issues prev`

Get the previous sibling sub-issue by position.

```sh
gh aux sub-issues prev --issue <parent-number> --sub-issue <target-number> [--repo OWNER/REPO]
```

### `sub-issues next`

Get the next sibling sub-issue by position.

```sh
gh aux sub-issues next --issue <parent-number> --sub-issue <target-number> [--repo OWNER/REPO]
```
