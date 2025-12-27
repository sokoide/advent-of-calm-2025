# Advent of CALM 2025 - Project Context

このプロジェクトは、FINOSの **Common Architecture Language Model (CALM)** を用いたアーキテクチャ・モデリングの学習と実践を目的としています。

## プロジェクト概要

CALMは、複雑なシステム（特に金融サービスやクラウドアーキテクチャ）を記述するための宣言的なJSONベースのモデリング言語です。このリポジトリでは、24日間の学習プロセスを通じて、ノード、リレーションシップ、フロー、コントロール（NFRs）、パターン、標準（Standards）などの概念を実装・検証しています。

### 主要コンポーネント
- **Go DSL (`go/`)**: CALMのJSONスキーマに準拠したモデルをプログラムから動的に生成するためのカスタムDSL。
- **Architectures (`architectures/`)**: 生成または手動作成されたCALMモデルファイル。
- **Patterns & Standards (`patterns/`, `standards/`)**: アーキテクチャのガバナンスと再利用性を高めるための定義ファイル。
- **Documentation (`docs/`)**: CALMモデルから生成されたドキュメント、ADR（Architectural Decision Records）、運用マニュアル。

## 技術スタック
- **Language**: Go 1.25.5
- **Framework**: FINOS CALM 1.0
- **Tools**:
  - `calm-cli`: バリデーション、生成、テンプレート処理。
  - `golines`: Goコードの整形（120文字制限）。
  - `jq`: JSONデータの比較・加工。

## 開発・ビルドコマンド
`go/` ディレクトリ内の `Makefile` を使用して主要な操作を行います。

- **ビルドと実行**:
  - `make build`: Goツールをビルドして `../arch-gen` を生成。
  - `make run`: アーキテクチャJSONを生成して標準出力に表示。
- **検証とテスト**:
  - `make validate`: `calm validate` を実行して、生成されたJSONがCALMスキーマに準拠しているか検証。
  - `make diff`: 生成されたJSONと `architectures/ecommerce-platform.json` の差異を確認。
- **コード品質**:
  - `make format`: `golines` を使用してコードを整形。
  - `make setup`: 必要なツール（golines）のインストール。

## 開発規約
- **Go DSLの利用**: 新しいノードやリレーションシップを追加する場合は、`go/main.go` 内の fluent API を使用してください。
- **スキーマ準拠**: すべての変更は `make validate` を通してCALM 1.0スキーマに準拠していることを確認する必要があります。
- **ドキュメント更新**: アーキテクチャの変更に伴い、`docs/` 下のドキュメントも（テンプレートやCLIツールを用いて）適宜更新してください。
- **ADRの参照**: `go/main.go` 内で ADR (`docs/adr/`) をモデルにリンクさせ、決定のトレーサビリティを維持してください。

## 特記事項
- **CALM VSCode Extension**: 編集中のリアルタイムプレビューには `Cmd+Shift+C` を使用します。
- **Chatmode**: `.github/chatmodes/CALM.chatmode.md` には、AIアシスタント向けのCALM特化型プロンプトが含まれています。
