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
| [`commits`](cmd/commits/README.md) | Query commit metadata (e.g. find the PR associated with a commit) |
| [`files`](cmd/files/README.md) | Get or download file contents from a remote repository at any ref |
| [`pr-comments`](cmd/prcomments/README.md) | List, get, update, reply to, and add pull request comments (inline and general) |
| [`projects`](cmd/projects/README.md) | Add issues/PRs to a Project V2 and set field values |
| [`sub-issues`](cmd/subissues/README.md) | List and manage sub-issues of a GitHub issue |
