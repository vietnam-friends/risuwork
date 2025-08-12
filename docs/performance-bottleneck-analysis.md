# RISUWORK パフォーマンスボトルネック分析レポート

## ISUCONについて

ISUCON（Iikanjini Speed Up Contest）は、与えられたWebアプリケーションを限られた時間内で高速化し、そのスコアを競う技術コンテストです。主な特徴：
- 8時間の競技時間内でパフォーマンスチューニングを実施
- ベンチマーカーが負荷走行してスコアを算出
- 計測→ボトルネック特定→改善のサイクルを高速に回すことが重要

## 1. 最重要ボトルネック（緊急対応必要）

### 1.1 N+1問題（最優先度：極高）

#### JobSearchService.java - 求人検索で深刻なN+1問題
**ファイル**: `webapp/java/src/main/java/jp/co/recruit/isucon2024/cs/api/service/JobSearchService.java`

```java
// 問題のコード（35-46行目）
for (JobEntity jobEntity : jobEntities) {
    // 各求人ごとに個別のSQLクエリを実行
    CompanyWithIndustryNameEntity entity = 
        jobSearchDao.selectCompanyWithIndustryNameByJobId(jobEntity.getId());
    // ...
}
```

**影響度**: 
- 検索結果100件の場合、101回のクエリ実行（1+100）
- レスポンスタイムが10倍以上に悪化する可能性

**改善案**:
```sql
-- 1回のクエリで全データ取得
SELECT j.*, c.id as company_id, c.name as company_name, 
       ic.name as industry_name, c.industry_id
FROM job j
JOIN user u ON j.create_user_id = u.id
JOIN company c ON u.company_id = c.id
JOIN industry_category ic ON c.industry_id = ic.id
WHERE j.id IN (?)
```

#### ApplicationsService.java - 応募履歴でN+1問題
**ファイル**: `webapp/java/src/main/java/jp/co/recruit/isucon2024/cs/api/service/ApplicationsService.java`

```java
// 問題のコード（38-41行目）
for (ApplicationWithJobEntity applicationWithJobEntity : applicationWithJobEntityList) {
    JobEntity jobEntity = applicationsDao.selectJobById(applicationWithJobEntity.getJob_id());
    // 各応募ごとに求人情報を個別取得
}
```

**影響度**: 応募履歴20件で21回のクエリ実行

## 2. データベース関連のボトルネック

### 2.1 インデックス不足（優先度：高）

現在のスキーマには重要なインデックスが一切定義されていません。

**必要なインデックス**:
```sql
-- 最重要インデックス
CREATE INDEX idx_user_email ON user(email);  -- ログイン処理の高速化
CREATE INDEX idx_job_search ON job(is_active, is_archived, updated_at DESC, id DESC);  -- 検索の高速化
CREATE INDEX idx_job_create_user_id ON job(create_user_id);  -- 企業の求人一覧取得
CREATE INDEX idx_application_user_id ON application(user_id);  -- 応募履歴取得
CREATE INDEX idx_user_company_id ON user(company_id);  -- 企業ユーザー検索

-- 追加で有効なインデックス
CREATE INDEX idx_job_salary ON job(salary);  -- 給与範囲検索
CREATE INDEX idx_application_job_id ON application(job_id);  -- 求人への応募一覧
CREATE INDEX idx_application_created_at ON application(created_at DESC);  -- 応募履歴のソート
```

### 2.2 非効率なクエリ構造

#### ネストしたサブクエリの問題
**ファイル**: `webapp/java/src/main/java/jp/co/recruit/isucon2024/cs/api/dao/JobSearchDao.java`

```sql
-- 現在の非効率なクエリ
WHERE company.id = 
  (SELECT company_id FROM user WHERE id = 
    (SELECT create_user_id FROM job WHERE id = ?))
```

**改善案**: JOINを使用した効率的なクエリに変更

### 2.3 タグ検索の非効率性
```java
// 4つのLIKE句で検索（インデックスが効かない）
AND (tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?)
```

**問題**: 
- LIKE検索でインデックスが効かない
- カンマ区切りでタグを保存しているため検索効率が悪い

**改善案**: タグを別テーブルに正規化

## 3. アプリケーション層のボトルネック

### 3.1 メモリ効率の問題（優先度：中）

#### 全件取得してからのページング
```java
// GetJobsService.java - 全件取得後にアプリ側でページング
List<JobEntity> jobEntityList = getJobsDao.selectNotArchivedJobsByCompanyId(userEntity.getCompany_id());
// その後、Javaコード内でページング処理
```

**問題**: 
- 大量データの場合、メモリを圧迫
- データベースから不要なデータを転送

**改善案**: 
```sql
-- LIMIT/OFFSETを使用
SELECT * FROM job 
WHERE is_archived = false 
ORDER BY updated_at DESC 
LIMIT ? OFFSET ?
```

### 3.2 セッション管理の非効率性

同一リクエスト内で同じユーザー情報を複数回取得している箇所が多数存在。

**改善案**: リクエストスコープでのキャッシュ実装

## 4. nginx設定のボトルネック

### 4.1 基本的な最適化設定の不足

現在の`nginx.conf`は最小限の設定のみ：

```nginx
# 現在の設定（改善が必要）
location / {
    proxy_pass http://localhost:8080;
}
```

**必要な最適化設定**:
```nginx
http {
    # 接続の再利用
    keepalive_timeout 65;
    keepalive_requests 100;
    
    # Gzip圧縮
    gzip on;
    gzip_types text/plain application/json text/css application/javascript;
    gzip_min_length 1000;
    
    # バッファサイズの最適化
    client_body_buffer_size 16K;
    client_header_buffer_size 1k;
    large_client_header_buffers 4 16k;
    
    # アップストリームの接続プール
    upstream backend {
        server localhost:8080;
        keepalive 32;
    }
    
    location / {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        
        # プロキシバッファの最適化
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
    
    # 静的ファイルのキャッシュ（もし存在すれば）
    location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
        expires 1d;
        add_header Cache-Control "public, immutable";
    }
}
```

## 5. 改善実施の優先順位

### Phase 1（即座に実施）- 期待効果：スコア5倍以上
1. **インデックスの追加**（30分）
   - 最も簡単で効果が高い
   - 特に`idx_user_email`と`idx_job_search`は必須

2. **N+1問題の解決**（2時間）
   - JobSearchServiceのN+1問題を解決
   - ApplicationsServiceのN+1問題を解決

### Phase 2（次に実施）- 期待効果：スコア2倍
3. **nginx設定の最適化**（30分）
   - Gzip圧縮の有効化
   - Keepaliveの設定
   - バッファサイズの調整

4. **データベースクエリの最適化**（1時間）
   - LIMIT/OFFSETの実装
   - サブクエリをJOINに変更

### Phase 3（時間があれば）- 期待効果：スコア1.5倍
5. **キャッシュの実装**（2時間）
   - セッション情報のキャッシュ
   - 業界カテゴリのキャッシュ

6. **タグ検索の正規化**（3時間）
   - タグテーブルの作成
   - データ移行とクエリ変更

## 6. 計測とモニタリング

### 推奨ツール
- **AWS X-Ray**: 既に統合済み、分散トレーシングに活用
- **スロークエリログ**: MySQLのslow_query_logを有効化
- **JProfiler/VisualVM**: Javaアプリケーションのプロファイリング

### 重要メトリクス
- クエリ実行回数（特にN+1問題の確認）
- 各APIエンドポイントのレスポンスタイム
- データベースのCPU使用率
- メモリ使用量

## まとめ

最も効果的な改善は以下の3つ：
1. **インデックスの追加**（実装容易、効果大）
2. **N+1問題の解決**（効果極大）
3. **nginx設定の最適化**（実装容易、効果中）

これらを実施するだけで、ベンチマークスコアは10倍以上向上する可能性があります。