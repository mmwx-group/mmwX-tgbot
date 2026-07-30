# Repository Guidelines

## Project Structure & Module Organization

This repository contains a Go 1.24 Telegram bot with no local database. `cmd/mmwx-tgbot/main.go` is the executable entry point. Application code lives under `internal/`: `bot/` contains command routing, handlers, rate limiting, notifications, and the Mini App HTTP server; `config/` loads YAML and `MMWX_TGBOT_*` environment variables; `mmwxclient/` provides typed access to the mmwX admin API. Mini App images and SVGs are in `internal/bot/assets/`. Deployment and release tooling lives in root shell scripts, `scripts/`, `Dockerfile`, `docker-compose.yml`, and `.github/workflows/`.

## Build, Test, and Development Commands

- `go run ./cmd/mmwx-tgbot -c ./config.yaml` runs the bot locally against a configured mmwX instance.
- `go build ./...` compiles all packages and catches build errors.
- `go vet ./...` runs the static checks required by CI.
- `go test ./...` runs all package tests; add tests as the suite grows.
- `./build.sh` creates the current-platform binary in `build/`; `./build.sh -a` cross-compiles all supported targets.
- `docker compose up -d` starts the containerized service.

Before submitting changes, run `go test ./...`, `go vet ./...`, and `go build ./...`.

## Coding Style & Naming Conventions

Use standard Go formatting (`gofmt -w`) and tabs as emitted by `gofmt`. Follow idiomatic Go naming: exported identifiers use `PascalCase`, internal identifiers use `camelCase`, and filenames use lowercase words with underscores where useful (for example, `handlers_admin_create.go`). Keep command handlers grouped in `handlers_*.go` and register new commands in `internal/bot/router.go`. Add endpoint-specific request/response types and methods to `internal/mmwxclient/tgbot_api.go`; keep generic HTTP behavior in `client.go`.

## Testing Guidelines

There are currently no committed `_test.go` files or explicit coverage threshold. Add focused table-driven tests beside the package under test, named `*_test.go`, with functions such as `TestConfigValidate`. Mock HTTP dependencies with `httptest` and avoid calls to real Telegram or mmwX services. Cover authorization, validation, and error paths when changing handlers or API clients.

## Commit & Pull Request Guidelines

Recent history uses terse Chinese feature summaries (for example, `支持公告`) and version-only release commits such as `v0.1.2`. Keep commits short, imperative, and limited to one logical change; reserve version messages for releases. Pull requests should explain behavior and configuration changes, list verification commands, link relevant issues, and include screenshots for Mini App UI changes.

## Security & Configuration

Copy `config.example.yaml` to `config.yaml` for local use. Never commit bot tokens, API tokens, production URLs, or Telegram IDs. Keep `webapp_dev_preview: false` in production and preserve server-side admin checks for every privileged route.
