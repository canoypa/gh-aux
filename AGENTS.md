# gh-aux

GitHub CLI extension that wraps GitHub operations — not available in standard `gh` — as named subcommands that are safe to auto-approve.

See [README.md](README.md) for installation and command list.

## Purpose and Direction

gh-aux wraps GitHub operations that require `gh api` or `gh graphql` into named subcommands. `gh api`/`gh graphql` cannot be selectively auto-approved; named subcommands can. Subcommand names and flag values align with GitHub UI concepts rather than raw API paths and node IDs.

**Guiding principle**: A command belongs in gh-aux when it is a recurring need in day-to-day agent workflows AND requires `gh api` or `gh graphql` to execute. Complexity and multi-step ID resolution are common reasons an operation isn't already in standard `gh`, but they are not requirements on their own.

## Build and Test

```sh
go build ./...
go test ./...
```

Requires Go 1.25+. Use `mise install` to set up the correct Go version.

To test locally:

```sh
go build -o gh-aux . && ./gh-aux <command-group> <subcommand>
```

To test as a gh extension:

```sh
gh extension install .
gh aux <command-group> <subcommand>
```

## Conventions

**Naming alignment with GitHub API**: Subcommand names and flag names should mirror the underlying GitHub API operation where possible. For example, `addProjectV2ItemById` → `projects add`, `deleteProjectV2Item` → `projects remove`, `updateProjectV2ItemFieldValue` → `projects update-field`, `clearProjectV2ItemFieldValue` → `projects clear-field`. Prefer the verb from the API mutation/endpoint over generic CRUD terms that don't match.

**UI-aligned naming over raw API paths**: The compound `<group> <subcommand>` name should read like an action a human would take in the GitHub UI (`projects update-field`, `sub-issues parent`, etc.).

### Adding a command group

1. Create `cmd/<group>/` as a new package (`package <group>`, all lowercase, no hyphens/underscores)
2. Files: `cmd.go` (registers the group), `types.go` (shared types and helpers), one file per subcommand. `types.go` contains per-group copies of `resolveRepo()` and `writeJSON()` — copied per group to keep groups independent (no shared internal package)
3. Register the group in `cmd/root.go`

### Output contract

- **stdout**: JSON only — a single object or array depending on the command
- **stderr**: Human-readable error messages (cobra default format is sufficient)
- **exit code**: 0 on success, non-zero on error
- **error handling**: Use `RunE`; return errors directly. Cobra writes them to stderr. Do not call `os.Exit` or `fmt.Fprintf(os.Stderr, ...)` manually

### API clients

- Use `api.DefaultGraphQLClient()` for GraphQL queries
- Use `api.DefaultRESTClient()` for REST calls
- Both come from `github.com/cli/go-gh/v2/pkg/api`
- REST POST/PATCH bodies: `json.Marshal` + `bytes.NewReader`

### Documentation

When adding a command group, also update:
- `README.md` Commands table (one row per group)
- `skills/gh-aux/SKILL.md` Command Reference table (one row per group)
- Add `skills/gh-aux/references/<group>.md` with commands, flags, output schema, and usage patterns
