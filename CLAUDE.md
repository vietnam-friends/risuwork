# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリで作業する際のガイダンスを提供します。

## 概要

RISUWORK (RECRUIT ISUCON 2024) は、クライアント（CL - 求人を投稿する企業）とカスタマー（CS - 求職者）向けに別々のAPIを持つ求人プラットフォームです。システムは複数の言語実装（Go、Java、Node.js）で構築されており、デプロイのニーズに応じて切り替えることができます。

## アーキテクチャ

アプリケーションは多層アーキテクチャに従っています：

1. **フロントエンドポータル** (`/portal`) - モニタリングと可視化のためのNext.jsアプリケーション
2. **Web API** (`/webapp`) - nginx背後の3つの実装（Go、Java、Node.js）
3. **ベンチマーカー** (`/benchmarker`) - Goベースの負荷テストと検証ツール
4. **インフラストラクチャ** (`/infrastructure`) - AWSデプロイ用のTerraform/OpenTofu設定

主要なアーキテクチャの決定事項：
- CLとCSユーザー両方のセッションベース認証
- 永続ストレージ用のMySQLデータベース
- 分散トレーシング用のAWS X-Ray統合
- ecspressoを使用したコンテナオーケストレーション用のECSデプロイ

## 共通開発コマンド

### Portal (Next.js)
```bash
cd portal
npm install          # 依存関係のインストール
npm run dev          # ポート3000で開発サーバーを起動
npm run build        # 本番用ビルド
npm run lint         # ESLintを実行
```

### Webアプリケーション - Go
```bash
cd webapp/go
go mod download      # 依存関係のダウンロード
go build main.go     # アプリケーションのビルド
go run main.go       # アプリケーションの実行
```

### Webアプリケーション - Java
```bash
cd webapp/java
mvn clean install    # ビルドと依存関係のインストール
mvn spring-boot:run  # アプリケーションの実行
mvn package          # JARファイルの作成
```

### Webアプリケーション - Node.js
```bash
cd webapp/nodejs
npm install          # 依存関係のインストール
npm run dev          # ホットリロード付き開発サーバーの起動
npm run build        # TypeScriptのコンパイル
npm start            # コンパイル済みアプリケーションの実行
```

### ベンチマーカー
```bash
cd benchmarker
go run ./cmd/bench/... run localhost:8080     # フルベンチマークの実行
go run ./cmd/lightbench/... localhost:8080    # 軽量ベンチマークの実行
```

### Dockerを使用したローカル開発
```bash
# Goバックエンド
cd webapp/local-docker/go-backend
docker compose up

# Javaバックエンド
cd webapp/local-docker/java-backend
docker compose up

# Node.jsバックエンド
cd webapp/local-docker/nodejs-backend
docker compose up
```

## データベース

MySQLデータベーススキーマは以下で初期化されます：
- `/webapp/sql/01_schema.sql` - データベーススキーマ
- `/webapp/sql/02_testdata.sql` - テストデータ

主要テーブル：
- `users` - タイプ区別を持つCLとCSの両ユーザー
- `jobs` - CLユーザーが作成した求人投稿
- `applications` - CSユーザーからの求人応募
- `companies` - CLユーザーの企業情報

## APIエンドポイント

アプリケーションは2つの主要なAPIグループを公開しています：

### CL（クライアント）API - `/cl/api/`
- 認証：サインアップ、ログイン、ログアウト
- 企業管理：企業作成
- 求人管理：作成、更新、アーカイブ

### CS（カスタマー）API - `/cs/api/`
- 認証：サインアップ、ログイン、ログアウト
- フィルタリング付き求人検索
- 求人応募
- 応募履歴

### 共通API - `/api/`
- ベンチマーク中のデータリセット用初期化エンドポイント

## テストと検証

ベンチマーカーは以下を検証します：
- APIレスポンスの正確性
- パフォーマンスメトリクス
- 操作間のデータ一貫性
- セッション管理

検証シナリオの実行：
```bash
cd benchmarker
go run ./cmd/bench/... run --validate-only localhost:8080
```

## デプロイ

GitHub ActionsワークフローがAWS ECSへのデプロイを処理します：
- `deploy-go` - Go実装のデプロイ
- `deploy-java` - Java実装のデプロイ
- `deploy-node` - Node.js実装のデプロイ

各ワークフロー：
1. Dockerイメージのビルド
2. ECRへのプッシュ
3. ecspresso経由でECSサービスの更新

## 重要な注意事項

- すべての言語実装はAPI互換性を維持する必要があります
- セッションクッキーは本番環境でセキュア設定を使用
- パフォーマンスモニタリング用にAWS X-Rayトレーシングが有効
- ベンチマーカーは特定のレスポンス形式を期待 - 変更はそれに対して検証
- データベース接続はパフォーマンスのためにコネクションプーリングを使用