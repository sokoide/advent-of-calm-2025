.PHONY: build format run validate diff clean help setup

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make setup    - 必要なツール (golines) をインストールします"
	@echo "  make build    - Go プログラムをビルドして実行ファイルを生成します"
	@echo "  make format   - golines を使用してソースコードを整形します (120文字制限)"
	@echo "  make run      - プログラムを実行してアーキテクチャ JSON を出力します"
	@echo "  make validate - アーキテクチャを生成し、CALM バリデーションを実行します"
	@echo "  make diff     - 既存の ecommerce-platform.json との差分を確認します"
	@echo "  make clean    - 生成されたバイナリや一時ファイルを削除します"

# セットアップ: 必要なツールをインストール
setup:
	go install github.com/segmentio/golines@latest

# ビルド: 実行ファイルを生成
build:
	go build -o ../arch-gen .

# フォーマット: golines で 120 文字制限を適用して整形
format:
	@command -v golines >/dev/null 2>&1 || (echo "golines が見つかりません。make setup を実行してください。" && exit 1)
	golines -w -m 120 .

# 実行: JSON を標準出力に書き出し
run:
	@go run .

# バリデーション: JSON を生成して calm validate を実行
validate:
	@go run . > generated-arch.json
	calm validate -a generated-arch.json
	@rm generated-arch.json

# 比較: 既存の ecommerce-platform.json との差分を確認
diff:
	@jq -S . ../architectures/ecommerce-platform.json > /tmp/expected.json
	@go run . > /tmp/raw_actual.json
	@jq -S . /tmp/raw_actual.json > /tmp/actual.json
	diff -u /tmp/expected.json /tmp/actual.json || true
	@rm /tmp/expected.json /tmp/actual.json /tmp/raw_actual.json

# クリーンアップ: 生成物を削除
clean:
	rm -f ../arch-gen generated-arch.json
