# Repository Guidelines

## Project Structure & Module Organization
`architectures/` hosts CALM JSON models for individual days and scenarios; start here when reviewing or extending a system definition. `patterns/` keeps reusable CALM governance patterns, whereas `standards/` captures node/relationship guardrails and `templates/` holds markdown shells for generated reports (e.g., `templates/ops/service-runbook.md`). `docs/` is the output directory for CALM-generated documentation, and the root `README.md` summarizes the journey. The Go DSL lives in `go/`, supported by `scripts/` helpers such as `validate.sh`, and generated binaries land beside the repository root when `make -C go build` runs.

## Build, Test, and Development Commands
- `make -C go build` — compiles `go/` into `../arch-gen`, the executable that emits CALM JSON.
- `make -C go run` or `go run .` inside `go/` — streams the current architecture JSON to stdout for spot checks.
- `make -C go format` — runs `golines -m 120 --base-formatter gofmt` to enforce the 120-character limit before commits.
- `make -C go validate` — writes `generated-arch.json`, runs `calm validate -a generated-arch.json`, and cleans up the temp file (requires the CALM CLI on `PATH`).
- `make -C go diff` — normalizes the existing `architectures/ecommerce-platform.json` and the newly generated output with `jq` before showing a `diff -u`.

## Coding Style & Naming Conventions
Go files use idiomatic formatting (tabs for indentation, `gofmt` output) with `golines` enforcing 120-character soft wraps. Factory methods favor noun names (e.g., `arch.Node()`, `node.Interface()`), constructors begin with `New…`, mutators use `Add…`, `Set…`, or `With…`, and fluent setters return the receiver. Files mirror their package role (e.g., `arch_dsl.go`, `helpers.go`) and live under `go/`; keep generated JSON under `architectures/` and never commit temporary build artifacts (`arch-gen`, `generated-arch.json`).

## Testing Guidelines
CALM JSON validation is the actionable test in this repo. Run `make -C go validate` after editing the DSL or architecture definitions; any failure means the JSON or patterns violate CALM schema expectations. There are no unit tests, so rely on architecture diff (`make -C go diff`) and spot checks from `calm validate -a architectures/<file>` when adding new models. Keep generated JSON filenames descriptive (use the day or scenario name) to align with the existing `architectures/` entries.

## Commit & Pull Request Guidelines
History shows short, focused commits (`refactor`, `adjust: 1 line length`), so prefer concise messages—optionally prefix with a type (`fix:`, `docs:`, `feat:`) followed by the substantive change. Pull requests should include: a summary of what changed, which CALM models or days are affected, any new validation steps taken, and links to related docs/screenshots (especially when the change adds new visualization content in `docs/` or `screenshots/`).

## Documentation & Configuration Notes
`docs/` is fully generated from the CALM CLI; regenerate it with `calm template` or via the VSCode extension when the underlying model changes. Keep `url-mapping.json` synchronized with new `docs/` pages so published links resolve correctly. Use `scripts/validate.sh` for quick sanity checks if you need to automate CLI validation outside the Go Makefile.
