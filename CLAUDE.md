# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is an **Advent of CALM 2025** project - a learning and practice repository for FINOS CALM (Common Architecture Language Model). The project implements a custom Go DSL and CALM Studio (React-based visual editor) with bidirectional synchronization between code and GUI.

**Core Concept**: Go DSL serves as the "Source of Truth" - GUI edits are translated to AST patches that rewrite the Go source code directly using `go/ast`.

## Development Environment

- **Go**: 1.25.5 darwin/arm64
- **Frontend**: React 19.2.0, TypeScript 5.9.3, Vite 7.2.4
- **Package Manager**: go mod, npm
- **Key Tools**: `calm-cli` (CALM validation), `d2` (diagrams), `golines` (formatting, 120 char limit)

## Build Commands

All commands must be run from the `go/` directory.

```bash
cd go/
```

### Primary Development Commands

| Command | Purpose |
|---------|---------|
| `make setup` | Install required tools (d2 via brew) |
| `make build` | Build arch-gen executable |
| `make format` | Format Go code with gofmt |
| `make run` | Generate architecture JSON to stdout |
| `make validate` | Generate JSON and run CALM validation |
| `make check` | Run Go DSL validation rules (no JSON output) |
| `make clean` | Remove build artifacts and generated files |

### Testing

```bash
make test               # Run all unit tests
make test-coverage      # Generate and show test coverage
make testcoverage       # Show coverage per package

# Running single tests (from go/ directory)
go test -v ./internal/infra/ast -run TestApplyPatch
go test -v ./internal/domain -run TestValidation
go test -v ./internal/usecase -run TestEcommerceBuilder
```

### Development Server and Visualization

```bash
make studio             # Launch CALM Studio (Editor + AI Agent)
make studio-ui          # Launch CALM Studio (Editor only)
make watch              # Live server with auto-refresh (Mermaid)
make watch-d2           # Live server with auto-refresh (D2)
make d2                 # Generate D2 diagram source and SVG
```

### Diff Tools

```bash
make diff               # Compare generated JSON with ecommerce-platform.json
make difftool           # Visual diff with nvim
make diff-arch          # Show semantic architecture difference
```

### Frontend Build

The frontend in `cmd/studio/frontend/` must be built before running `make studio`:

```bash
cd cmd/studio/frontend
npm install
npm run build
```

## Project Architecture

### High-Level Structure

```
go/
├── cmd/                        # Entry points
│   ├── arch-gen/              # Architecture generator (main CLI tool)
│   ├── studio/                # CALM Studio (backend + embedded React frontend)
│   ├── arch-agent/            # AI agent for architecture assistance
│   ├── diff/                  # Architecture diff tool
│   └── watch/                 # Live preview server
├── internal/
│   ├── domain/                # Core domain models
│   │   ├── architecture.go         # Architecture model structure
│   │   ├── patch.go                # Patch operation types (15+ types)
│   │   ├── origin.go               # Code origin tracking
│   │   └── validation.go           # Validation rules
│   ├── infra/                 # Infrastructure layer
│   │   ├── ast/                   # AST manipulation (patch_impl.go: ~2000 lines)
│   │   ├── d2/                    # D2 diagram rendering
│   │   ├── filesystem/            # File I/O
│   │   ├── parser/                # Parser utilities
│   │   ├── render/                # JSON/D2/Mermaid rendering
│   │   └── generator/             # DSL execution
│   └── usecase/                # Business logic
│       ├── ecommerce_architecture.go  # Reference architecture (~617 lines)
│       └── code_sync.go          # Code synchronization
└── Makefile
```

### Bidirectional Synchronization

**Go DSL → GUI (Rendering)**:
1. `internal/infra/generator` executes DSL (`EcommerceBuilder.Build()`)
2. Constructs `domain.Architecture` model in memory
3. Converts to JSON and sends to Studio frontend
4. React Flow renders nodes/edges from JSON

**GUI → Go DSL (Patching)**:
1. User action in GUI (rename, move, delete) → JSON patch operation
2. `domain.PatchOperation` sent to backend
3. `internal/infra/ast.GoASTSyncer` parses Go source using `go/ast`
4. Rewrites AST based on patch content
5. Formats and writes back to file

### Key Files to Read

| File | Purpose |
|------|---------|
| `go/docs/ARCH.md` | Complete architecture documentation with AST manipulation details and Mermaid diagrams |
| `go/internal/domain/patch.go` | Patch operation types and structs |
| `go/internal/infra/ast/patch_impl.go` | Core AST patch implementation (2014 lines) |
| `go/internal/usecase/ecommerce_architecture.go` | Reference architecture definition |

### Patch Operation Types

Defined in `domain/patch.go`:

| Type | Description |
|------|-------------|
| `add-node` / `update-node` / `delete-node` | Node CRUD operations |
| `add-relationship` / `update-relationship` / `delete-relationship` | Connection operations |
| `add-interface` / `delete-interface` | Interface operations |
| `add-flow` / `update-flow` / `delete-flow` | Flow operations |
| `add-composed-of` / `update-composed-of` / `delete-composed-of` | Container operations |
| `add-control` / `update-control` / `delete-control` | Control operations |
| `update-count` | Loop variable updates |

### AST Manipulation Patterns

The `internal/infra/ast` package uses `go/ast` for precise code manipulation:

**Node Identification**:
- Priority: Line number from `Origin.Line`
- Fallback: Search for `DefineNode` calls with matching ID argument

**Property Updates**:
- Basic args: Update `Args[2]` (name) and `Args[3]` (description)
- Functional options: Scan variadic `Args[4:]` for `WithOwner(...)` calls

**Deletion with Cascade**:
- Identify variable name from assignment
- Delete statement containing `DefineNode`
- Search and delete all references (relationships, composed-of) using variable name

Reference: `go/docs/ARCH.md` contains detailed Mermaid diagrams showing AST structure.

## DSL Naming Conventions

| Prefix/Method | Role |
|---------------|------|
| `New...` | Independent part generation |
| `Define...` | Declarative creation (uses Functional Options) |
| `With...` | Option setting for `Define...` methods |
| `ConnectTo` | Node-centric connection |
| `Via` / `Is` / `Encrypted` | Attribute setting |

## Development Patterns

### Adding a New Patch Operation

1. Add `PatchType` constant to `internal/domain/patch.go`
2. Add fields to `PatchOperation` struct if needed
3. Add case to `ApplyPatch()` in `internal/infra/ast/patch_impl.go`
4. Implement helper functions (`addXxxToAST`, `updateXxxInAST`)
5. Add tests to `patch_impl_test.go`

### Adding Validation Rules

1. Define rule implementing `ValidationRule` interface in `internal/domain/validation.go`
2. Create test cases in `validation_test.go`

## Testing

- **Domain tests**: `internal/domain/*_test.go`
- **Infrastructure tests**: `internal/infra/*_test.go`
- **Use case tests**: `internal/usecase/*_test.go`

Run tests with verbose output to see individual test results:

```bash
go test -v ./internal/infra/ast -run TestApplyPatch
```

## Known Limitations

1. **Complex Control Flow**: Changes inside loops/conditionals have limitations
2. **Type Updates**: Node type changes require import awareness (deferred)
3. **Comment Preservation**: `go/format` may shift comment positions
4. **AST Position Tracking**: Line number matching can be fragile

## CALM CLI Usage

```bash
# Install globally
npm install -g @finos/calm-cli

# Validate architecture
calm validate -a architectures/ecommerce-platform.json

# Generate from pattern
calm generate -p patterns/api-gateway.json -o architecture.json

# Generate documentation
calm docify -a architecture.json -o docs/
```

## Serena MCP Project Memory

This project uses Serena MCP for session persistence. Memories are stored in `.serena/memories/`:

- `project_summary.md` - Project overview
- `style_and_conventions.md` - Coding conventions
- `suggested_commands.md` - Common commands
- `on_task_completion.md` - Task completion checklist

Activate with: `/sc:load` or use Serena tools directly.

---

**Last Updated**: 2026-01-26
