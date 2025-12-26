# CALM Standards

このディレクトリには、CALM (Common Architecture Language Model) アーキテクチャに適用される組織固有の標準仕様が含まれています。

## 1. Standards とは何か
CALM Standards は、コアとなる CALM 仕様の上に独自の制約や属性を拡張するための仕組みです。

### allOf Composition による拡張
アーキテクチャ定義内では、各コンポーネントがコアの仕様と組織固有の標準の両方を満たす必要があることを **`allOf`** 構成によって宣言します。これにより、CALM 本来の柔軟性を維持しつつ、ガバナンスに必要な項目を必須化できます。

## 2. ノード標準要件 (`company-node-standard.json`)
すべてのノードは、以下の組織固有プロパティを持つ必要があります：

*   **costCenter**: `CC-####` 形式（例: `CC-1234`）のコストセンターID。
*   **owner**: ノードの責任を持つチームまたは個人の名前。
*   **environment**: 展開環境。以下のいずれかである必要があります：
    *   `development`
    *   `staging`
    *   `production`

## 3. リレーションシップ標準要件 (`company-relationship-standard.json`)
コンポーネント間のすべての接続は、以下の項目を明示する必要があります：

*   **dataClassification**: 通信されるデータの感度レベル。以下のいずれかである必要があります：
    *   `public`
    *   `internal`
    *   `confidential`
    *   `restricted`
*   **encrypted**: 通信が暗号化されているかを示す真偽値 (`true` または `false`)。

## 使用例
アーキテクチャファイルでの参照例：
```json
{
  "nodes": [
    {
      "unique-id": "my-service",
      "standards": ["https://example.com/standards/company-node-standard.json"],
      "costCenter": "CC-9999",
      "owner": "App-Team",
      "environment": "production",
      ...
    }
  ]
}
```
