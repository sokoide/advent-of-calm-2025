.PHONY: build format run validate diff difftool clean help setup watch diff-arch d2 watch-d2 check studio test test-coverage testcoverage studio-ui

# Default target
help:
	@echo "Available commands:"
	@echo "  make setup     - Install required tools (golines)"
	@echo "  make build     - Build the Go program and generate executable"
	@echo "  make format    - Format source code using golines (120 char limit)"
	@echo "  make run       - Run the program and output architecture JSON"
	@echo "  make validate  - Generate architecture and run CALM validation"
	@echo "  make check     - Run Go DSL validation rules"
	@echo "  make test      - Run unit tests"
	@echo "  make test-coverage - Check test coverage"
	@echo "  make testcoverage - Show test coverage per package"
	@echo "  make diff      - Check diff with existing ecommerce-platform.json"
	@echo "  make difftool  - Visually check diff with existing ecommerce-platform.json"
	@echo "  make diff-arch - Show semantic difference of the architecture"
	@echo "  make d2        - Generate D2 diagrams"
	@echo "  make watch     - Start live server with auto-refresh (Mermaid)"
	@echo "  make watch-d2  - Start live server with auto-refresh (D2)"
	@echo "  make studio    - Launch CALM Studio (Editor + AI Agent)"
	@echo "  make studio-ui - Launch CALM Studio (Editor only)"
	@echo "  make clean     - Remove generated binaries and temporary files"

# Setup: Install required tools
setup:
	go install github.com/segmentio/golines@latest
	@echo "Install D2: brew install d2"

# Build: Generate executable
build:
	go build -o ../arch-gen ./cmd/arch-gen

# Format: Apply 120 character limit using golines
format:
	@command -v golines >/dev/null 2>&1 || (echo "golines not found. Run make setup." && exit 1)
	golines -w -m 120 --base-formatter gofmt .

# Run: Output JSON to stdout
run:
	@go run ./cmd/arch-gen

# Validate: Generate JSON and run calm validate
validate:
	@go run ./cmd/arch-gen > generated-arch.json
	calm validate -a generated-arch.json
	@rm generated-arch.json

# Go DSL Validation: Run built-in rules
check:
	@go run ./cmd/arch-gen -validate

# Test: Run unit tests
test:
	go test ./...

# Coverage: Check test coverage
test-coverage:
	GOCACHE=$(PWD)/.gocache go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Show coverage per package
testcoverage:
	@for pkg in $$(GOCACHE=$(PWD)/.gocache go list ./...); do \
		echo "==> $$pkg"; \
		GOCACHE=$(PWD)/.gocache go test -cover $$pkg; \
	done

# Diff: Check difference with existing ecommerce-platform.json
diff:
	@jq -S . ../architectures/ecommerce-platform.json > /tmp/expected.json
	@go run ./cmd/arch-gen > /tmp/raw_actual.json
	@jq -S . /tmp/raw_actual.json > /tmp/actual.json
	diff -u /tmp/expected.json /tmp/actual.json || true
	@rm /tmp/expected.json /tmp/actual.json /tmp/raw_actual.json

# Visual Diff: Compare with existing ecommerce-platform.json using nvim
difftool:
	@jq -S . ../architectures/ecommerce-platform.json > /tmp/expected.json
	@go run ./cmd/arch-gen > /tmp/raw_actual.json
	@jq -S . /tmp/raw_actual.json > /tmp/actual.json
	nvim -d /tmp/expected.json /tmp/actual.json
	@rm /tmp/expected.json /tmp/actual.json /tmp/raw_actual.json

# Cleanup: Remove artifacts
clean:
	rm -f ../arch-gen generated-arch.json architecture.d2 coverage.out
	rm -rf .gocache

# Live Server: Watch file changes and auto-refresh browser (Mermaid)
watch:
	@echo "🚀 Starting live server (Mermaid)..."
	@cd cmd/watch && go run . ../..

# Architecture Diff: Show semantic difference
diff-arch:
	@go run ./cmd/arch-gen > /tmp/new-arch.json
	@cd cmd/diff && go run . ../../../architectures/ecommerce-platform.json /tmp/new-arch.json

# D2 Diagram Generation
d2:
	@go run ./cmd/arch-gen -format d2 > architecture.d2
	@echo "✅ Generated architecture.d2"
	@command -v d2 >/dev/null 2>&1 && d2 architecture.d2 architecture.svg && echo "✅ Generated architecture.svg" || echo "💡 Install d2 to generate SVG: brew install d2"

# D2 Live Server: Watch file changes and auto-refresh D2 diagrams
watch-d2:
	@echo "🚀 Starting D2 live server..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 not found. Run brew install d2." && exit 1)
	@cd cmd/watch && go run . -d2 ../..

# CALM Studio: Editor + AI Agent (Full set)
studio: cmd/studio/frontend/dist
	@echo "🧭 Starting local agent + CALM Studio..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 not found. Run brew install d2." && exit 1)
	@bash -c 'set -euo pipefail; \
		(go run ./cmd/arch-agent -dir . -port 8787 2>&1 | sed -u "s/^/agent: /") & AGENT_PID=$$!; \
		(STUDIO_GENERATE_MODE=gorun go run ./cmd/studio 2>&1 | sed -u "s/^/studio: /") & STUDIO_PID=$$!; \
		trap "kill $$AGENT_PID $$STUDIO_PID 2>/dev/null || true" INT TERM EXIT; \
		wait $$AGENT_PID $$STUDIO_PID'

# CALM Studio UI: Bidirectional Editor only (for offline/manual editing)
studio-ui: cmd/studio/frontend/dist
	@echo "🎨 Starting CALM Studio UI..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 not found. Run brew install d2." && exit 1)
	@cd cmd/studio && STUDIO_GENERATE_MODE=gorun go run . ../..

cmd/studio/frontend/dist: cmd/studio/frontend/src/App.tsx
	@echo "📦 Building frontend..."
	@cd cmd/studio/frontend && npm install && npm run build