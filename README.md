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

| Command                                           | Description                                                       |
| ------------------------------------------------- | ----------------------------------------------------------------- |
| [`commits`](cmd/commits/README.md)                | Query commit metadata (e.g. find the PR associated with a commit) |
| [`review-comments`](cmd/reviewcomments/README.md) | List, get, update, and add inline pull request review comments    |
| [`reviews`](cmd/reviews/README.md)                | Create and submit pull request reviews                            |
| [`projects`](cmd/projects/README.md)              | Add issues/PRs to a Project V2 and set field values               |
