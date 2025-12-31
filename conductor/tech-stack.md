# Technology Stack

## Core Language & Runtime
- **Go 1.25.5**: プロジェクトの主要開発言語。型安全なDSLの構築に使用。

## Architecture Modeling
- **FINOS CALM 1.0**: アーキテクチャ記述の標準スキーマ。
- **Custom Go DSL**: Functional Options Patternを用いた、宣言的なモデル定義。

## Tools & Utilities
- **React Flow**: インタラクティブなダイアグラムエディタの基盤。
- **Monaco Editor (@monaco-editor/react)**: ブラウザ上での高度なコード編集環境。
- **React Resizable Panels**: リサイズ可能な分割レイアウトの実現。
- **Go AST (go/ast)**: Go ソースコードの解析・自動書き換えエンジン。
- **D2 CLI**: サーバーサイドでの SVG レンダリングおよび D2 DSL 生成。
- **Go AST (Sync Engine)**: JSON から Go DSL への安全な同期を実現するカスタム AST トランスフォーマ。
- **Vite / React / Tailwind CSS 4**: モダンなフロントエンド開発スタック。
- **calm-cli**: 生成されたJSONのスキーマバリデーションおよびテンプレート処理。
- **Go Test**: `make test` によるユニットテスト実行と、`make test-coverage` によるカバレッジ計測。
- **D2**: 建築図面の高度なレンダリングエンジン。
- **Mermaid**: 軽量な図面レンダリング（Live Preview用）。
- **golines**: 120文字制限を遵守するためのGoコード整形ツール。
- **jq**: JSONデータの加工・比較。

## Architecture & Patterns
- **Clean Architecture**: Domain, Usecase, Infraの3層構造による責任の分離。
- **Ports & Adapters**: Interfaceによる外部依存の抽象化。
