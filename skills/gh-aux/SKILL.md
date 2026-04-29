---
name: gh-aux
description: gh-aux gives agents a stable named interface for GitHub operations that would otherwise require raw `gh api` or `gh graphql` calls with multi-step ID resolution. Use when the task involves Project V2 field operations, PR review thread access, or sub-issue management.
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

| Command group        | Description                                    | Reference                                |
| -------------------- | ---------------------------------------------- | ---------------------------------------- |
| `gh aux pr-comments` | Manage PR comments (inline + general)          | [pr-comments](references/pr-comments.md) |
| `gh aux projects`    | Add issues/PRs to Project V2, set field values | [projects](references/projects.md)       |
| `gh aux sub-issues`  | List and manage sub-issues of a GitHub issue   | [sub-issues](references/sub-issues.md)   |
