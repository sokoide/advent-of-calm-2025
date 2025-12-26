# My Advent of CALM Journey

This repository tracks my 24-day journey learning the Common Architecture Language Model (CALM).

## Progress

- [x] Day 1: Install CALM CLI and Initialize Repository
- [x] Day 2: Create Your First Node
- [x] Day 3: Connect Nodes with Relationships
- [x] Day 4: Install CALM VSCode Extension
- [x] Day 5: Add Interfaces to Nodes
- [x] Day 6: Document with Metadata

## Tools

### CALM CLI

CALM CLI は、CALM ファイルの生成、検証、テンプレート処理を行うためのコマンドラインツールです。

- **主な用途**:
  - `generate`: パターンからアーキテクチャの雛形を生成
  - `validate`: アーキテクチャやパターンが仕様に準拠しているか検証
  - `template`: CALM モデルからドキュメントやレポートを生成
- **基本コマンド**:

    ```bash
    # バリデーションの実行
    calm validate -a architectures/my-first-architecture.json

    # ヘルプの表示
    calm --help
    ```

### CALM VSCode Extension

VSCode 内で CALM アーキテクチャを効率的に編集・閲覧するためのプラグインです。

- **Marketplace**: [FINOS CALM VSCode Plugin](https://marketplace.visualstudio.com/items?itemName=FINOS.calm-vscode-plugin)
- **主な機能**:
  - **Visualization**: アーキテクチャ図のリアルタイムプレビュー
  - **Tree Navigation**: ノードやリレーションシップの構造をツリー形式で表示
  - **Live Preview**: 編集内容を即座に図に反映
- **ショートカット**: `Ctrl+Shift+C` / `Cmd+Shift+C` でプレビュー画面を開きます。

### 連携方法

これらのツールを組み合わせることで、効率的な開発サイクルを実現できます：

- **CLI**: CI/CD パイプラインやコミット前の自動チェック（バリデーション）に使用します。
- **Extension**: 開発中の視覚的なフィードバックや、複雑な構造の理解、ドキュメント作成時の図の確認に使用します。

## Architectures

This directory will contain CALM architecture files documenting systems.

## Patterns

This directory will contain CALM patterns for architectural governance.

## Docs

Generated documentation from CALM models.
