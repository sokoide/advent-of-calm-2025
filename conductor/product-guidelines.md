# Product Guidelines

このガイドラインは、Go 1.25.5環境における「Architecture as Code」の品質を維持するためのものです。

## 1. Go Development Standards ($go-master)
- **Idiomatic Go**: 標準ライブラリを優先し、明確で読みやすいコードを記述します。
- **Functional Options Pattern**: 複雑な構造体（Node, Relationship等）の初期化には、拡張性と可読性に優れたFunctional Optionsを使用します。
- **Strict Formatting**: `golines` を使用し、120文字の行制限を遵守します。
- **Error Handling**: センチネルエラーやカスタムエラー型を適切に使用し、コンテキスト情報を付与して上位層へ伝播させます。

## 2. Clean Architecture Standards ($ca-master)
- **Layer Separation**:
    - `domain`: 外部（JSONスキーマ等）に依存しない純粋なアーキテクチャモデル。
    - `usecase`: モデルを操作し、DSLのロジックを実現するアプリケーション層。
    - `infra`: JSON/D2のパース・レンダリングなど、具体的な実装詳細。
- **Dependency Inversion**: 上位層（Domain/Usecase）は常にInterface（Port）に依存し、具体的な実装（Adapter）には依存しません。
- **Package Integrity**: パッケージ間の循環参照を厳禁とし、責任の境界を明確にします。

## 3. Testing & TDD Principles
- **Test First**: 可能な限り、新しい機能やバリデーションルールを追加する前にテストを記述します。
- **Mocking**: インフラ層との境界ではInterfaceを定義し、モックを使用してドメインロジックの純粋なテストを可能にします。
- **Validation as Test**: CALMモデルの整合性チェック自体をテストコードの一部として扱います。

## 4. Documentation & Decision Tracking
- **SSoT (Single Source of Truth)**: アーキテクチャ情報はGo DSLで一元管理し、JSONやドキュメントはそこから生成します。
- **ADR Integration**: 重要な設計上の決定は ADR として記録し、コード内のコメントやメタデータからリンクさせます。
