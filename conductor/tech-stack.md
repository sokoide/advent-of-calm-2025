# Technology Stack

## Core Language & Runtime
- **Go 1.25.5**: プロジェクトの主要開発言語。型安全なDSLの構築に使用。

## Architecture Modeling
- **FINOS CALM 1.0**: アーキテクチャ記述の標準スキーマ。
- **Custom Go DSL**: Functional Options Patternを用いた、宣言的なモデル定義。

## Tools & Utilities
- **calm-cli**: 生成されたJSONのスキーマバリデーションおよびテンプレート処理。
- **Go Test**: `make test` によるユニットテスト実行と、`make test-coverage` によるカバレッジ計測。
- **D2**: 建築図面の高度なレンダリングエンジン。
- **Mermaid**: 軽量な図面レンダリング（Live Preview用）。
- **golines**: 120文字制限を遵守するためのGoコード整形ツール。
- **jq**: JSONデータの加工・比較。

## Architecture & Patterns
- **Clean Architecture**: Domain, Usecase, Infraの3層構造による責任の分離。
- **Ports & Adapters**: Interfaceによる外部依存の抽象化。
