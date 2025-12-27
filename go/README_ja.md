# CALM Architecture Definition in Go

このディレクトリには、CALM アーキテクチャを Go で定義・生成するための DSL 実装が含まれています。JSON を手書きするのではなく、Go の型安全性を活かした「アーキテクチャのコーディング」を目的としています。

## 主な改善点 (Recent Improvements)

- **Functional Options パターン**: `DefineNode` において `WithOwner` や `WithMeta` を使った宣言的な定義が可能になりました。
- **メタデータの部品化 (Composition)**: `Merge` ヘルパーにより、共通のメタデータ部品（Tier1, DBA, ManagedService等）を組み合わせてノードを構築できます。
- **Fluent Connection API**: `node.ConnectTo(dest)` 形式の直感的な配線と、`LinksContainer` による関係の構造化管理を導入しました。
- **型安全なフロー構築**: 文字列 ID への依存を減らし、リレーションシップオブジェクトやその `GetID()` メソッドを介してフローを定義できます。
- **`owner` の一元管理 (SSoT)**: `WithOwner` オプションにより、トップレベルの `owner` フィールドと `metadata["owner"]` の両方が自動的に同期されます。
- **動的コンポーネント対応**: ゲートウェイの数や ID を変更しても、依存関係 (`dependencies`) やフロー定義が自動的に追随します。

## アーキテクチャの一貫性と `make diff`

最新の DSL では、アーキテクチャモデルの一貫性を高めるため、すべてのノードにおいて `owner` 情報をトップレベルとメタデータの両方に自動セットします。

これに合わせて、リポジトリ内の参照用 JSON (`architectures/ecommerce-platform.json`) も Go DSL の出力結果で更新されました。これにより、モデル全体の品質と一貫性が向上しています。

以前の（一部欠落があった）JSON と比較した際の diff の例：

```diff
     {
       "unique-id": "customer",
       "name": "Customer",
+      "metadata": {
+        "owner": "marketing-team"
+      },
       "owner": "marketing-team"
     }
```

## DSL 命名規則 (Naming Conventions)

| 接頭辞 / メソッド | 役割 | 説明 | 例 |
| :--- | :--- | :--- | :--- |
| **`New...`** | **独立した部品の生成** | 親が決まっていない単体の部品を作成します。 | `NewRequirement`, `NewSecurityConfig` |
| **`Define...`** | **宣言的生成 (Modern)** | Functional Options を受け取り、高度に構成されたオブジェクトを生成します。 | `arch.DefineNode()`, `arch.DefineFlow()` |
| **`With...`** | **オプション設定** | `Define...` メソッドに渡すための設定関数です。 | `WithOwner()`, `WithMeta()`, `WithControl()` |
| **`ConnectTo`** | **ノード中心の接続** | ノード自身から接続を開始し、Builder を返します。 | `node.ConnectTo(dest)` |
| **`Via` / `Is` / `Is`** | **属性の設定 (Fluent)** | オブジェクトのプロパティを流れるように設定します。 | `rel.Via("src", "dst").Encrypted(true)` |
| **`Merge`** | **メタデータの合成** | 複数のマップを一つにまとめます。キー衝突時はパニックします。 | `Merge(metaTier1, metaOps)` |

## 推奨されるコーディングパターン

### 1. メタデータの部品化

共通の設定を変数として定義し、必要に応じてマージします。

```go
metaOps = map[string]any{"oncall": "#oncall-ops"}
a.DefineNode("svc", Service, "Name", "Desc", WithMeta(Merge(metaTier1, metaOps)))
```

### 2. リレーションシップの構造化管理

`LinksContainer` 構造体ですべての接続を保持し、フロー定義で参照します。

```go
lc.OrderToInv = n.OrderSvc.ConnectTo(n.InventorySvc, "Check stock").WithID("order-rel")
// フローで参照
fb.Step(lc.OrderToInv.GetID(), "Checking inventory")
```

## 使用方法

```bash
# フォーマット (120文字制限適用)
make -C go format

# バリデーション (CALM スキーマチェック)
make -C go validate

# オリジナル JSON との比較 (diff が出なければ成功)
make -C go diff
```
