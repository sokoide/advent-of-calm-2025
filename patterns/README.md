# CALM Patterns

このディレクトリには、共通の構成を定義し強制するための CALM (Common Architecture Language Model) パターンが含まれています。

## 1. パターンとは何か
CALM パターンは、アーキテクチャ設計における**「生成 (Generation)」**と**「バリデーション (Validation)」**という2つの強力な役割（スーパーパワー）を持っています。

*   **生成**: パターンをテンプレートとして使用し、標準に準拠した新しいアーキテクチャファイルを即座に生成できます。
*   **バリデーション**: 作成されたアーキテクチャが組織の標準構造（ノード数、接続の向き、プロトコルなど）に従っているかを厳格にチェックします。

## 2. Web アプリケーションパターン (`web-app-pattern.json`)
このパターンは、標準的な 3層 Web アプリケーションの基本構造を定義しています。

### 使用方法

#### アーキテクチャの生成
パターンから新しいアーキテクチャファイルの雛形を作成します。
```bash
calm generate -p patterns/web-app-pattern.json -o architectures/my-new-app.json
```

#### アーキテクチャのバリデーション
既存のファイルがパターンに適合しているか検証します。
```bash
calm validate -p patterns/web-app-pattern.json -a architectures/generated-webapp.json
```

## 3. パターンによる強制事項
このパターンは以下の構造を厳格に要求します：
*   **ノード (計3つ)**:
    1.  `web-frontend` (node-type: webclient)
    2.  `api-service` (node-type: service)
    3.  `app-database` (node-type: database)
*   **リレーションシップ (計2つ)**:
    1.  `frontend-to-api`: Frontend から API への接続
    2.  `api-to-database`: API から Database への接続

## 4. カスタマイズ可能な柔軟性
パターンの制約を満たした上で、以下の項目は自由に拡張可能です：
*   **詳細説明**: 各ノードやリレーションシップの `description`
*   **インターフェース**: ホスト名、ポート、プロトコルの詳細設定
*   **メタデータ**: 所有者、リポジトリURL、運用情報、タグ
*   **コントロール**: セキュリティ、パフォーマンス、可用性の要件
