# Contributing

Thanks for helping improve `Rubin Investment`.

## Before You Start

- Please read [README.md](README.md) and [docs/runtime-configuration.md](docs/runtime-configuration.md).
- Do not commit personal databases, private cookies, tokens, or machine-specific paths.
- If a change depends on private runtime overrides, document the public fallback behavior.

## Development Flow

1. Install dependencies for Go and `frontend/`.
2. Run `go test ./...`.
3. Run `npm run build` inside `frontend/`.
4. Keep changes scoped and explain user-visible behavior in the pull request.

## Pull Requests

- Keep PRs focused on one concern when possible.
- Add or update tests when behavior changes.
- Mention any new runtime configuration fields or required local setup.
- Avoid checking in generated local data from `data/`, `logs/`, or export folders.

## Security and Secrets

- Never paste secrets into public issues or pull requests.
- If you find a leak or a vulnerability, follow [SECURITY.md](SECURITY.md).

## Attribution

This repository includes upstream-derived `go-stock` code. Please preserve existing attribution and license notices when modifying reused upstream files.
