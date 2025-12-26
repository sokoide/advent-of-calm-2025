# CALM Architecture Definition in Go

このディレクトリには、CALM (Common Architecture Language Model) アーキテクチャを Go 言語で定義・生成するための高度な DSL (Domain Specific Language) 実装が含まれています。

## Go DSL 方式の主なメリット

巨大な JSON ファイルを直接編集する代わりに Go コードを使用することで、設計の品質と保守性が劇的に向上します。

### 1. 宣言的な Fluent API

`NewNode(...).WithInterfaces(...).WithControl(...)` といったメソッドチェーン（Fluent API）を採用しています。これにより、システムのトポロジーや制約を自然言語に近い形で、かつ極めて簡潔に記述できます。

### 2. コンパイル時チェック（型安全）

* **構造化された Config**: `SecurityConfig` や `CircuitBreakerConfig` などの専用構造体を使用するため、プロパティ名のタイポや型の不一致をコンパイル時に検出できます。
* **列挙型の活用**: `NodeType` などの重要な識別子が定数化されており、無効な値の混入を未然に防ぎます。

### 3. メタデータと共通設定の再利用

* **Metadata ビルダー**: `NewMetadata().Add(key, value)` により、柔軟なメタデータ定義を IDE の補完を受けながら行えます。
* **一括適用**: 共通の責任者情報や SLA 設定を変数化し、複数のノードやリレーションシップに一括して適用することが容易です（DRY 原則の徹底）。

### 4. 運用ナレッジの統合

障害モード (`failure-modes`) や監視リンク (`monitoring`) をコードの近くに定義することで、アーキテクチャ図を「単なる設計図」から「生きた運用仕様書」へと進化させることができます。

### 5. 自動化されたワークフロー

付属の `Makefile` により、フォーマット、ビルド、生成、そして CALM 仕様への適合チェック（バリデーション）をワンコマンドで実行可能です。

## ファイル構成

* `arch_dsl.go`: CALM 1.1 仕様に基づいたコア構造体定義。
* `helpers.go`: Fluent API を実現するコンストラクタとビルダーメソッド。
* `main.go`: アーキテクチャの具体的な定義（ビジネスロジック）。
* `Makefile`: 開発サイクルの自動化コマンド。

## 使用方法

ソースコードの整形からバリデーションまでを自動で行います：

```bash
# コードの整形
make -C go format

# アーキテクチャの生成とバリデーション
make -C go validate

# 特定のファイルに出力する場合
go run go/*.go > architectures/ecommerce-platform.json
```
