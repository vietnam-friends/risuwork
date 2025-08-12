# RISUWORK API エンドポイント仕様書

## 概要

RISUWORK は求人プラットフォームのAPIサービスです。企業（CL: Client）と求職者（CS: Customer）の2つのユーザータイプをサポートし、求人の投稿・検索・応募機能を提供します。

各言語実装（Go, Java, Node.js）で同一のAPIインターフェースを提供しており、nginx経由でアクセスされます。

## 認証方式

- セッションベース認証（Cookie使用）
- セッション有効期限：3600秒（1時間）
- Cookie設定：
  - HttpOnly: true
  - SameSite: Strict
  - Path: /

## 共通仕様

### ページネーション

- ページ番号は0ベースインデックス
- ページサイズ：
  - 求人検索（CS）: 50件/ページ
  - 応募一覧（CS）: 20件/ページ
  - 求人一覧（CL）: 50件/ページ

### エラーレスポンス

| HTTPステータス | 説明 |
|--------------|------|
| 400 | Bad Request - リクエストパラメータ不正 |
| 401 | Unauthorized - 未認証 |
| 403 | Forbidden - 権限なし |
| 404 | Not Found - リソースが存在しない |
| 409 | Conflict - リソースの競合（重複登録など） |
| 422 | Unprocessable Entity - 処理不可能なエンティティ |
| 500 | Internal Server Error - サーバーエラー |

## APIエンドポイント一覧

### 1. 共通API

#### 1.1 初期化
```
POST /api/initialize
```

**説明**: ベンチマーカー用API。データベースを初期化し、テストデータを投入する。

**リクエスト**: なし

**レスポンス**:
```json
{
  "lang": "go" // または "java", "nodejs"
}
```

#### 1.2 終了処理
```
POST /api/finalize
```

**説明**: ベンチマーカー用API。終了時の処理を実行する。

**リクエスト**: なし

**レスポンス**: "ok"

---

### 2. CS（Customer/求職者）API

#### 2.1 サインアップ
```
POST /api/cs/signup
```

**説明**: 求職者アカウントを新規作成する。

**リクエスト**:
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "山田太郎"
}
```

**レスポンス**:
```json
{
  "message": "CS account created successfully",
  "id": 1234
}
```

**エラー**:
- 409: メールアドレスが既に使用されている

#### 2.2 ログイン
```
POST /api/cs/login
```

**説明**: 求職者としてログインする。

**リクエスト**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**レスポンス**: "Logged in successfully"

**エラー**:
- 401: メールアドレスまたはパスワードが不正

#### 2.3 ログアウト
```
POST /api/cs/logout
```

**説明**: ログアウトしてセッションを破棄する。

**認証**: 必要

**リクエスト**: なし

**レスポンス**: "Logged out successfully"

#### 2.4 求人検索
```
GET /api/cs/job_search
```

**説明**: 条件に基づいて求人を検索する。アーカイブ済みと非アクティブな求人は除外される。

**リクエストパラメータ**:
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| keyword | string | × | タイトルまたは説明文でのキーワード検索 |
| min_salary | int | × | 最低給与（以上） |
| max_salary | int | × | 最高給与（以下） |
| tag | string | × | タグによる絞り込み（完全一致） |
| industry_id | string | × | 業種IDによる絞り込み |
| page | int | × | ページ番号（0ベース、デフォルト: 0） |

**レスポンス**:
```json
{
  "jobs": [
    {
      "id": 1,
      "title": "バックエンドエンジニア",
      "description": "Goを使用したAPI開発",
      "salary": 6000000,
      "tags": "Go,API,Backend",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "company": {
        "id": 100,
        "name": "株式会社テック",
        "industry": "IT・通信"
      }
    }
  ],
  "page": 0,
  "has_next_page": true
}
```

**ソート順**: updated_at DESC, id DESC

#### 2.5 求人応募
```
POST /api/cs/application
```

**説明**: 指定した求人に応募する。

**認証**: 必要（CSユーザーのみ）

**リクエスト**:
```json
{
  "job_id": 123
}
```

**レスポンス**:
```json
{
  "message": "Successfully applied for the job",
  "id": 456
}
```

**エラー**:
- 403: CLユーザーでのアクセス
- 404: 求人が存在しない
- 409: 既に応募済み
- 422: 求人が応募を受け付けていない（非アクティブまたはアーカイブ済み）

#### 2.6 応募一覧取得
```
GET /api/cs/applications
```

**説明**: ログインユーザーの応募履歴を取得する。

**認証**: 必要

**リクエストパラメータ**:
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| page | int | × | ページ番号（0ベース、デフォルト: 0） |

**レスポンス**:
```json
{
  "applications": [
    {
      "id": 456,
      "job_id": 123,
      "user_id": 789,
      "created_at": "2024-01-02T00:00:00Z",
      "job": {
        "id": 123,
        "title": "フロントエンドエンジニア",
        "description": "React/TypeScriptを使用した開発",
        "salary": 5500000,
        "tags": "React,TypeScript,Frontend",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    }
  ],
  "page": 0,
  "has_next_page": false
}
```

**ソート順**: created_at DESC

---

### 3. CL（Client/企業）API

#### 3.1 企業登録
```
POST /api/cl/company
```

**説明**: 新規企業を登録する。

**リクエスト**:
```json
{
  "name": "株式会社サンプル",
  "industry_id": "IT-001"
}
```

**レスポンス**:
```json
{
  "message": "Company created successfully",
  "id": 100
}
```

#### 3.2 サインアップ
```
POST /api/cl/signup
```

**説明**: 企業アカウントを新規作成する。

**リクエスト**:
```json
{
  "email": "hr@company.com",
  "password": "password123",
  "name": "人事担当者",
  "company_id": 100
}
```

**レスポンス**:
```json
{
  "message": "Signed up successfully",
  "id": 200
}
```

**エラー**:
- 400: 存在しない企業ID
- 409: メールアドレスが既に使用されている

#### 3.3 ログイン
```
POST /api/cl/login
```

**説明**: 企業アカウントとしてログインする。

**リクエスト**:
```json
{
  "email": "hr@company.com",
  "password": "password123"
}
```

**レスポンス**: "Logged in successfully"

**エラー**:
- 401: メールアドレスまたはパスワードが不正、またはCLユーザーではない

#### 3.4 ログアウト
```
POST /api/cl/logout
```

**説明**: ログアウトしてセッションを破棄する。

**認証**: 必要

**リクエスト**: なし

**レスポンス**: "Logged out successfully"

#### 3.5 求人作成
```
POST /api/cl/job
```

**説明**: 新規求人を作成する。

**認証**: 必要（CLユーザーのみ）

**リクエスト**:
```json
{
  "title": "バックエンドエンジニア募集",
  "description": "Goを使用したマイクロサービス開発",
  "salary": 6000000,
  "tags": "Go,Docker,Kubernetes"
}
```

**レスポンス**:
```json
{
  "message": "Job created successfully",
  "id": 123
}
```

**備考**: 作成時は `is_active: true`、`is_archived: false` で登録される

#### 3.6 求人更新
```
PATCH /api/cl/job/:jobid
```

**説明**: 既存の求人情報を部分更新する。

**認証**: 必要（求人作成企業の社員のみ）

**パスパラメータ**:
- jobid: 更新対象の求人ID

**リクエスト**（すべてオプション）:
```json
{
  "title": "新しいタイトル",
  "description": "新しい説明",
  "salary": 7000000,
  "tags": "新しいタグ",
  "is_active": false
}
```

**レスポンス**: "Job updated successfully"

**エラー**:
- 403: 他社の求人へのアクセス
- 404: 求人が存在しない
- 422: アーカイブ済みの求人

#### 3.7 求人アーカイブ
```
POST /api/cl/job/:jobid/archive
```

**説明**: 求人をアーカイブする。アーカイブ後は検索結果に表示されなくなる。

**認証**: 必要（求人作成企業の社員のみ）

**パスパラメータ**:
- jobid: アーカイブ対象の求人ID

**リクエスト**: なし

**レスポンス**: "Job archived successfully"

**エラー**:
- 403: 他社の求人へのアクセス
- 404: 求人が存在しない
- 422: 既にアーカイブ済み

**備考**: アーカイブ後も `/api/cl/job/:jobid` では取得可能

#### 3.8 求人詳細取得
```
GET /api/cl/job/:jobid
```

**説明**: 求人の詳細情報と応募者一覧を取得する。

**認証**: 必要（求人作成企業の社員のみ）

**パスパラメータ**:
- jobid: 取得対象の求人ID

**レスポンス**:
```json
{
  "id": 123,
  "title": "バックエンドエンジニア",
  "description": "Goを使用した開発",
  "salary": 6000000,
  "tags": "Go,API",
  "is_active": true,
  "create_user_id": 200,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "applications": [
    {
      "id": 456,
      "job_id": 123,
      "created_at": "2024-01-02T00:00:00Z",
      "applicant": {
        "id": 789,
        "email": "applicant@example.com",
        "name": "応募者名"
      }
    }
  ]
}
```

**エラー**:
- 403: 他社の求人へのアクセス
- 404: 求人が存在しない

**備考**: アーカイブ済みの求人も取得可能

#### 3.9 求人一覧取得
```
GET /api/cl/jobs
```

**説明**: 自社の求人一覧を取得する。アーカイブ済みの求人は除外される。

**認証**: 必要（CLユーザーのみ）

**リクエストパラメータ**:
| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| page | int | × | ページ番号（0ベース、デフォルト: 0） |

**レスポンス**:
```json
{
  "jobs": [
    {
      "id": 123,
      "title": "バックエンドエンジニア",
      "description": "Goを使用した開発",
      "salary": 6000000,
      "tags": "Go,API",
      "is_active": true,
      "create_user_id": 200,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "page": 0,
  "has_next_page": false
}
```

**ソート順**: updated_at DESC, id

---

## データベーススキーマ関連情報

### ユーザータイプ
- `CS`: Customer（求職者）
- `CL`: Client（企業担当者）

### 求人ステータス
- `is_active`: 応募受付中かどうか
- `is_archived`: アーカイブ済みかどうか

### タグ形式
- カンマ区切りで複数タグを格納
- 例: "Go,Docker,Kubernetes"

### 業種カテゴリ
- `industry_category` テーブルで管理
- 企業登録時に `industry_id` で指定

## 実装間の互換性

Go、Java、Node.jsの3つの実装はすべて同一のAPIインターフェースを提供しています。各実装で以下の点が共通です：

1. エンドポイントURL
2. HTTPメソッド
3. リクエスト/レスポンスのJSON構造
4. HTTPステータスコード
5. 認証方式（セッションベース）
6. ページネーション仕様
7. ソート順