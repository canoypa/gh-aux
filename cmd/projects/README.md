# projects

Add issues and pull requests to a GitHub Project V2, and set field values.

## Commands

### `projects add`

Add an issue or pull request to a Project V2.

```sh
gh aux projects add --project <number> --issue <number> [--field "Name=Value"] [--repo OWNER/REPO]
gh aux projects add --project <number> --pr <number> [--field "Name=Value"] [--repo OWNER/REPO]
```

`--field` can be repeated to set multiple fields in one call.

### `projects remove`

Remove an issue or pull request from a Project V2.

```sh
gh aux projects remove --project <number> --issue <number> [--repo OWNER/REPO]
gh aux projects remove --project <number> --pr <number> [--repo OWNER/REPO]
```

### `projects update-field`

Set a field value on an existing Project V2 item.

```sh
gh aux projects update-field --project <number> --issue <number> --field "Name=Value" [--repo OWNER/REPO]
gh aux projects update-field --project <number> --pr <number> --field "Name=Value" [--repo OWNER/REPO]
```

### `projects clear-field`

Clear a field value on a Project V2 item.

```sh
gh aux projects clear-field --project <number> --issue <number> --field-name "FieldName" [--repo OWNER/REPO]
gh aux projects clear-field --project <number> --pr <number> --field-name "FieldName" [--repo OWNER/REPO]
```
