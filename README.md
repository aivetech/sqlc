# sqlc AIVE

## Overview

sqlc AIVE doesn't check everything upfront. To ensure SQL quality, we need to create plugins that enforce quality standards such as:

- Checking that all mandatory columns are filled
- Ensuring specific columns are always mandatory
- Verifying indexing strategies
- And more...

## Creating a Plugin

To create a new plugin:

1. Create a new directory in `src/`
2. Initialize a Go module with: `go mod init plugin-name`
3. Implement your plugin logic

```
package plugin-name

func main() {
	codegen.Run(generate)
}

func generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	return &plugin.GenerateResponse{}, nil
}

```

## Releasing

To release a new version, create a pull request with a commit message starting with one of:

- `BREAKING` - for breaking changes
- `feat` - for new features
- `fix` - for bug fixes

The release will be automated based on the commit message prefix.

