# MySQL名前付きロックテスト

このプロジェクトは、MySQLの名前付きロック機能をテストするためのサンプルアプリケーションです。APIサーバーがDBに名前付きロックを取得し、各リクエストでロックされたセッションIDの内容を確認することができます。

## 機能

- MySQLの名前付きロック（GET_LOCK, RELEASE_LOCK）を使用したロック機構
- 複数のクライアントからのリクエストでロックの競合を確認
- 各セッションのセッションID（CONNECTION_ID）の取得
- ロックの所有者（IS_USED_LOCK）の確認
- ロックの状態（IS_FREE_LOCK）の確認
- ロック履歴の記録

## 技術スタック

- Go言語
- MySQL 8.0（Docker）
- 依存性注入：github.com/samber/do
- HTTPルーティング：github.com/labstack/echo/v4
- データベース：database/sql（標準ライブラリ）

## プロジェクト構成

```
named-lock/
├── cmd/
│   └── main.go                # アプリケーションのエントリーポイント
├── docker/
│   └── mysql/
│       └── init/
│           └── 01_init.sql    # MySQLの初期化スクリプト
├── internal/
│   ├── config/
│   │   └── config.go          # アプリケーション設定
│   ├── db/
│   │   └── db.go              # データベース操作
│   ├── handler/
│   │   └── lock_handler.go    # HTTPハンドラ
│   └── service/
│       └── lock_service.go    # ビジネスロジック
├── docker-compose.yml         # Docker Compose設定
├── test_client.go             # テスト用クライアント
└── README.md                  # このファイル
```

## セットアップと実行方法

### 1. MySQLコンテナの起動

```bash
docker-compose up -d
```

### 2. APIサーバーの起動

```bash
go run cmd/main.go
```

### 3. テストクライアントの実行

別のターミナルを開いて、以下のコマンドを実行します。引数にクライアントIDと並列数を指定できます。

```bash
# クライアント1を実行
go run test_client.go 1

# 別のターミナルでクライアント2を実行
go run test_client.go 2

# 複数のクライアントを並列実行（例: ID 1から始まる5つのクライアント）
go run test_client.go 1 5
```

並列実行の場合、第1引数は開始クライアントID、第2引数は並列数を指定します。各クライアントは独自のIDを持ち、並行してロックの取得・解放を試みます。

## APIエンドポイント

### セッションID取得

```
GET /api/session
```

レスポンス例:
```json
{
  "session_id": "123456"
}
```

### ロック取得

```
POST /api/locks
```

リクエスト例:
```json
{
  "lock_name": "test_lock",
  "timeout": 10
}
```

レスポンス例:
```json
{
  "success": true,
  "session_id": "123456",
  "message": "Lock acquired successfully. Current connection ID: 123456"
}
```

### ロック解放

```
DELETE /api/locks/{lockName}
```

レスポンス例:
```json
{
  "success": true,
  "session_id": "123456",
  "message": "Lock released successfully"
}
```

### ロック状態確認

```
GET /api/locks/{lockName}
```

レスポンス例:
```json
{
  "lock_name": "test_lock",
  "is_locked": true,
  "owner_session_id": "123456",
  "current_session_id": "123456",
  "is_owned_by_current_session": true
}
```

## テストシナリオ

1. クライアント1がロックを取得
2. クライアント2がロックの取得を試みる（失敗する）
3. クライアント1がロックを解放
4. クライアント2が再度ロックの取得を試みる（成功する）

このシナリオにより、MySQLの名前付きロックの動作と、各セッションIDの確認ができます。

## 注意点

- このプロジェクトはテスト・デモ用であり、本番環境での使用は想定していません。
- ロックのタイムアウト値は適切に設定してください。長すぎるとロックが解放されずに残る可能性があります。
- サーバー停止時にはロックは自動的に解放されますが、アプリケーションの不具合でロックが解放されない場合は、MySQLクライアントから手動で解放する必要があります。

## 手動でのロック操作（MySQLクライアント）

MySQLクライアントから直接ロックを操作する場合:

```sql
-- 現在のセッションIDを確認
SELECT CONNECTION_ID();

-- ロックを取得（第2引数はタイムアウト秒数）
SELECT GET_LOCK('test_lock', 10);

-- ロックの所有者を確認
SELECT IS_USED_LOCK('test_lock');

-- ロックが解放されているか確認
SELECT IS_FREE_LOCK('test_lock');

-- ロックを解放
SELECT RELEASE_LOCK('test_lock');
```
