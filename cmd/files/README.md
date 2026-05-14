# files

Get file contents from a remote repository at any ref.

## Commands

| Command                 | Description                                       | Required flags |
| ----------------------- | ------------------------------------------------- | -------------- |
| `gh aux files get`      | Get file contents or directory listing at a given ref | `--path`       |
| `gh aux files download` | Download a file from a repository to a local path | `--path`, `--output` |

## Flags

| Flag                | Description                                         | Default                        |
| ------------------- | --------------------------------------------------- | ------------------------------ |
| `--repo OWNER/REPO` | Target repository                                   | Current directory's git remote |
| `--path <path>`     | File path within the repository (e.g. `src/foo.ts`) | —                              |
| `--ref <ref>`       | Branch, tag, or commit SHA                          | Repository default branch      |
| `--output` / `-o`   | Local path to write the file to (`download` only)   | — (required)            |

## Output

- `get` (file) → `{ path, ref, sha, content, url }` — `content` is the full decoded file text
- `get` (directory) → `[{ type, name, path, sha, size, url }]` — `type` is `"file"` or `"dir"`
- `download` → `{ path, output, ref, sha, url }` — `output` is the absolute local path written

## Examples

```sh
# Get file contents as JSON
gh aux files get --path src/foo.ts
gh aux files get --path db/schema.rb --ref main

# Download a file
gh aux files download --path db/schema.rb --ref main --output /tmp/schema.rb

# Download to a specific path
gh aux files download --path db/schema.rb --output /tmp/schema.rb

# Another repository
gh aux files get --repo owner/repo --path README.md
gh aux files download --repo owner/repo --path README.md -o /tmp/README.md
```
