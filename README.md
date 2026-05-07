# gh-aux

[GitHub CLI](https://cli.github.com/) extension for AI agents. Wraps `gh api`/`gh graphql` operations into named subcommands that can be individually auto-approved.

## Installation

```sh
gh extension install canoypa/gh-aux
```

## Usage

```sh
gh aux <command> [subcommand] [flags]
```

## Commands

| Command | Description |
|---|---|
| [`pr-comments`](cmd/prcomments/README.md) | List, get, reply to, and add pull request comments (inline and general) |
| [`projects`](cmd/projects/README.md) | Add issues/PRs to a Project V2 and set field values |
| [`sub-issues`](cmd/subissues/README.md) | List and manage sub-issues of a GitHub issue |
