# CALM Architecture Definition in Go

このディレクトリには、CALM アーキテクチャを Go で定義・生成するための DSL 実装が含まれています。

## DSL 命名規則 (Naming Conventions)

コードの読みやすさと「書き心地」を維持するため、以下のルールに従ってメソッドを命名しています。

| 接頭辞 | 役割 | 説明 | 例 |
| :--- | :--- | :--- | :--- |
| **`New...`** | **独立した部品の生成** | 親が決まっていない単体の部品を作成します。 | `NewRequirement`, `NewSecurityConfig` |
| **(なし/名詞)** | **工場 (Factory)** | 親オブジェクトから子を生成し、自動的に親のリストへ追加します。 | `arch.Node()`, `node.Interface()`, `arch.Flow()` |
| **`Add...`** | **コレクションへの追加** | Metadata や Controls などのマップ/スライスへ要素を追加し、親を返します。 | `node.AddMeta()`, `node.AddControl()` |
| **`Set...`** / **`With...`** | **属性の設定 (Fluent)** | オブジェクト自身のプロパティ値を設定・変更し、自身を返します。 | `intf.SetPort()`, `rel.WithProtocol()` |

## 使用方法

```bash
# フォーマット (120文字制限適用)
make -C go format

# バリデーション
make -C go validate

# オリジナル JSON との比較
make -C go diff
```