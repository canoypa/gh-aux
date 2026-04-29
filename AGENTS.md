# gh-aux

GitHub CLI extension for GitHub operations that would otherwise require raw `gh api` or `gh graphql` calls.

See [README.md](README.md) for installation and command list.

## Purpose and Direction

gh-aux wraps complex GitHub API operations into named, predictable subcommands. The target users are agents and scripts that need structured output.

**Guiding principle**: A command belongs in gh-aux when the equivalent operation would require `gh api` or `gh graphql` with non-trivial request construction or response transformation.

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

Follow the naming style of existing commands for subcommand names and flags.

**Naming alignment with GitHub API**: Subcommand names and flag names should mirror the underlying GitHub API operation where possible. For example, `addProjectV2ItemById` → `add`, `deleteProjectV2Item` → `remove`, `updateProjectV2ItemFieldValue` → `update-field`, `clearProjectV2ItemFieldValue` → `clear-field`. Prefer the verb from the API mutation/endpoint over generic CRUD terms that don't match.

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
