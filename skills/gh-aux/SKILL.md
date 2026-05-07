---
name: gh-aux
description: gh-aux wraps `gh api` and `gh graphql` operations that require multi-step ID resolution or complex GraphQL into named subcommands. When you are about to write a `gh api` or `gh graphql` call, check gh-aux first.
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
| `gh aux pr-comments` | timeline, get, add, review, reply-review, resolve-thread | [pr-comments](references/pr-comments.md) |
| `gh aux projects`    | add, remove, update-field, clear-field                   | [projects](references/projects.md)       |
| `gh aux sub-issues`  | list, add, remove, prev, next, parent                    | [sub-issues](references/sub-issues.md)   |
