---
name: gh-aux
description: Use instead of raw `gh api` or `gh graphql` for GitHub workflows involving: PR review comments (add, reply, resolve), PR reviews (create/submit), Projects V2 (add/remove items, update fields), sub-issues, fetching file contents by ref, or finding a PR from a commit.
---

# gh-aux

GitHub CLI extension that provides named subcommands for GitHub operations that would otherwise require raw `gh api` or `gh graphql` calls.

## Setup

```sh
gh extension install canoypa/gh-aux
```

## Usage

```sh
gh aux <command-group> <subcommand> [flags]
```

`--repo` defaults to the current directory's git remote when omitted.

## Command Reference

| Command group        | Subcommands                                              | Reference                                |
| -------------------- | -------------------------------------------------------- | ---------------------------------------- |
| `gh aux commits`     | pr                                                       | [commits](references/commits.md)         |
| `gh aux files`       | get, download                                            | [files](references/files.md)             |
| `gh aux projects`    | add, remove, update-field, clear-field                   | [projects](references/projects.md)       |
| `gh aux review-comments` | list, get, update, add, resolve-thread               | [review-comments](references/review-comments.md) |
| `gh aux reviews`     | create                                                   | [reviews](references/reviews.md)         |
| `gh aux sub-issues`  | list, add, remove, prev, next, parent                    | [sub-issues](references/sub-issues.md)   |
