.PHONY: build format run validate diff difftool clean help setup watch diff-arch d2 watch-d2 check studio test test-coverage testcoverage studio-ui

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make setup     - 必要なツール (golines) をインストールします"
	@echo "  make build     - Go プログラムをビルドして実行ファイルを生成します"
	@echo "  make format    - golines を使用してソースコードを整形します (120文字制限)"
	@echo "  make run       - プログラムを実行してアーキテクチャ JSON を出力します"
	@echo "  make validate  - アーキテクチャを生成し、CALM バリデーションを実行します"
	@echo "  make check     - Go DSL のバリデーションルールを実行します"
	@echo "  make test      - ユニットテストを実行します"
	@echo "  make test-coverage - テストカバレッジを確認します"
	@echo "  make testcoverage - パッケージごとのテストカバレッジを表示します"
	@echo "  make diff      - 既存の ecommerce-platform.json との差分を確認します"
	@echo "  make difftool  - 既存の ecommerce-platform.json との差分を目で確認します"
	@echo "  make diff-arch - アーキテクチャの意味的な差分を表示します"
	@echo "  make d2        - D2 ダイアグラムを生成します"
	@echo "  make watch     - ライブサーバーを起動してブラウザで自動更新 (Mermaid)"
	@echo "  make watch-d2  - ライブサーバーを起動してブラウザで自動更新 (D2)"
	@echo "  make studio    - CALM Studio (エディタ + AIエージェント) を起動します"
	@echo "  make studio-ui - CALM Studio (エディタのみ) を起動します"
	@echo "  make clean     - 生成されたバイナリや一時ファイルを削除します"

# セットアップ: 必要なツールをインストール
setup:
	go install github.com/segmentio/golines@latest
	@echo "D2のインストール: brew install d2"

# ビルド: 実行ファイルを生成
build:
	go build -o ../arch-gen ./cmd/arch-gen

# フォーマット: golines で 120 文字制限を適用して整形
format:
	@command -v golines >/dev/null 2>&1 || (echo "golines が見つかりません。make setup を実行してください。" && exit 1)
	golines -w -m 120 --base-formatter gofmt .

# 実行: JSON を標準出力に書き出し
run:
	@go run ./cmd/arch-gen

# バリデーション: JSON を生成して calm validate を実行
validate:
	@go run ./cmd/arch-gen > generated-arch.json
	calm validate -a generated-arch.json
	@rm generated-arch.json

# Go DSL バリデーション: 内蔵ルールを実行
check:
	@go run ./cmd/arch-gen -validate

# テスト: ユニットテストを実行
test:
	go test ./...

# カバレッジ: テストカバレッジを確認
test-coverage:
	GOCACHE=$(PWD)/.gocache go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# パッケージごとのカバレッジを表示
testcoverage:
	@for pkg in $$(GOCACHE=$(PWD)/.gocache go list ./...); do \
		echo "==> $$pkg"; \
		GOCACHE=$(PWD)/.gocache go test -cover $$pkg; \
	done

# 比較: 既存の ecommerce-platform.json との差分を確認
diff:
	@jq -S . ../architectures/ecommerce-platform.json > /tmp/expected.json
	@go run ./cmd/arch-gen > /tmp/raw_actual.json
	@jq -S . /tmp/raw_actual.json > /tmp/actual.json
	diff -u /tmp/expected.json /tmp/actual.json || true
	@rm /tmp/expected.json /tmp/actual.json /tmp/raw_actual.json

# 比較: 既存の ecommerce-platform.json との差分を確認
difftool:
	@jq -S . ../architectures/ecommerce-platform.json > /tmp/expected.json
	@go run ./cmd/arch-gen > /tmp/raw_actual.json
	@jq -S . /tmp/raw_actual.json > /tmp/actual.json
	nvim -d /tmp/expected.json /tmp/actual.json
	@rm /tmp/expected.json /tmp/actual.json /tmp/raw_actual.json

# クリーンアップ: 生成物を削除
clean:
	rm -f ../arch-gen generated-arch.json architecture.d2 coverage.out
	rm -rf .gocache

# ライブサーバー: ファイル変更を監視してブラウザを自動更新 (Mermaid)
watch:
	@echo "🚀 Starting live server (Mermaid)..."
	@cd cmd/watch && go run . ../..

# アーキテクチャ差分: 意味的な差分を表示
diff-arch:
	@go run ./cmd/arch-gen > /tmp/new-arch.json
	@cd cmd/diff && go run . ../../../architectures/ecommerce-platform.json /tmp/new-arch.json

# D2 ダイアグラム生成
d2:
	@go run ./cmd/arch-gen -format d2 > architecture.d2
	@echo "✅ Generated architecture.d2"
	@command -v d2 >/dev/null 2>&1 && d2 architecture.d2 architecture.svg && echo "✅ Generated architecture.svg" || echo "💡 Install d2 to generate SVG: brew install d2"

# D2 ライブサーバー: ファイル変更を監視してD2ダイアグラムを自動更新
watch-d2:
	@echo "🚀 Starting D2 live server..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 が見つかりません。brew install d2 を実行してください。" && exit 1)
	@cd cmd/watch && go run . -d2 ../..

# CALM Studio: エディタ + AIエージェント (フルセット)
studio: cmd/studio/frontend/dist
	@echo "🧭 Starting local agent + CALM Studio..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 が見つかりません。brew install d2 を実行してください。" && exit 1)
	@bash -c 'set -euo pipefail; \
		(go run ./cmd/arch-agent -dir . -port 8787 2>&1 | sed -u "s/^/agent: /") & AGENT_PID=$$!; \
		(STUDIO_GENERATE_MODE=gorun go run ./cmd/studio 2>&1 | sed -u "s/^/studio: /") & STUDIO_PID=$$!; \
		trap "kill $$AGENT_PID $$STUDIO_PID 2>/dev/null || true" INT TERM EXIT; \
		wait $$AGENT_PID $$STUDIO_PID'

# CALM Studio UI: 双方向エディタのみ (オフライン/マニュアル編集用)
studio-ui: cmd/studio/frontend/dist
	@echo "🎨 Starting CALM Studio UI..."
	@command -v d2 >/dev/null 2>&1 || (echo "❌ d2 が見つかりません。brew install d2 を実行してください。" && exit 1)
	@cd cmd/studio && STUDIO_GENERATE_MODE=gorun go run . ../..

cmd/studio/frontend/dist: cmd/studio/frontend/src/App.tsx
	@echo "📦 Building frontend..."
	@cd cmd/studio/frontend && npm install && npm run build
