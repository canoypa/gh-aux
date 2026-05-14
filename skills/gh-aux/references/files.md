# files

Get file contents from a remote repository at any ref.

## Commands

| Command                 | Description                                           | Required flags       |
| ----------------------- | ----------------------------------------------------- | -------------------- |
| `gh aux files get`      | Get file contents or directory listing at a given ref | `--path`             |
| `gh aux files download` | Download a file from a repository to a local path     | `--path`, `--output` |

## Flags

| Flag                | Description                                         | Default                        |
| ------------------- | --------------------------------------------------- | ------------------------------ |
| `--repo OWNER/REPO` | Target repository                                   | Current directory's git remote |
| `--path <path>`     | File path within the repository (e.g. `src/foo.ts`) | —                              |
| `--ref <ref>`       | Branch, tag, or commit SHA                          | Repository default branch      |
| `--output` / `-o`   | Local path to write the file to (`download` only)   | — (required)                   |

## Output

- `get` (file) → `{ path, ref, sha, content, url }`
  - `content` — full decoded file content as a string
  - `ref` — the resolved ref (echoes `--ref` value, or `"HEAD"` when omitted)
- `get` (directory) → `[{ type, name, path, sha, size, url }]`
  - `type` — `"file"` or `"dir"`
- `download` → `{ path, output, ref, sha, url }`
  - `output` — absolute local path where the file was written

## Usage Patterns

**Get a file from the default branch (JSON output):**

```sh
gh aux files get --path src/foo.ts
```

**List a directory:**

```sh
gh aux files get --path cmd
gh aux files get --path cmd --ref main
```

**Get a file from a specific branch or commit:**

```sh
gh aux files get --path db/schema.rb --ref main
gh aux files get --path config/routes.rb --ref abc1234
```

**Download a file to a specific local path:**

```sh
gh aux files download --path db/schema.rb --output /tmp/schema.rb
```

**Get or download a file from another repository:**

```sh
gh aux files get --repo owner/repo --path README.md
gh aux files download --repo owner/repo --path README.md -o /tmp/README.md
```
