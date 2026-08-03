`ai-commit` is a CLI tool written in Go that generates Conventional Commit messages from staged Git diffs using the Gemini API.

## Requirements

- Go 1.25 or later
- Git
- Gemini API key

## Build

Compile the binaries and deploy to `~/dotfiles/bin/`:

```sh
./build.sh
```

Or build manually for local testing:

```sh
go build -o ai-commit .
```

## Configuration

Set your Gemini API key in your environment or in a `.env` file placed in the same directory as the executable:

```sh
GEMINI_API_KEY=your_api_key_here
```

Optional configuration:

- `AI_COMMIT_MODEL`: Model name (defaults to `gemini-1.5-flash`).
- `prompt.md`: If present in the executable directory, its text overrides the default system prompt.

## Usage

Stage changes in a Git repository and run the executable:

```sh
git add .
./ai-commit
```

To print the generated prompt without calling the API or committing, run with `--test-prompt` (or `-t`, `--dry-run`):

```sh
./ai-commit --test-prompt
```

When a commit message is proposed, select an action:

- `a`: Accept the message and execute `git commit`
- `r`: Reject the message and exit
- `c`: Enter feedback to regenerate the message

Staged diff generation automatically excludes lockfiles (`go.sum`, `package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`) and minified assets (`*.min.js`, `*.min.css`).

## Test

Run the test suite:

```sh
go test ./...
```
