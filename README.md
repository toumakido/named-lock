# MySQL名前付きロックテスト

このプロジェクトは、MySQLの名前付きロック機能をテストするためのサンプルアプリケーションです。APIサーバーがDBに名前付きロックを取得し、各リクエストでロックされたセッションIDの内容を確認することができます。

## 機能

- MySQLの名前付きロック（GET_LOCK, RELEASE_LOCK）を使用したロック機構
- 複数のクライアントからのリクエストでロックの競合を確認
- 各セッションのセッションID（CONNECTION_ID）の取得
- ロックの取得・保持・解放を一連の操作として実行する機能
- トランザクション内でのFOR UPDATE句を使用したデータ処理
- ロックを取得し、データを処理し、解放する一連の操作を実行する機能

## 技術スタック

- Go言語 (1.24.2)
- MySQL 8.0（Docker）
- 依存性注入：github.com/samber/do v1.6.0
- HTTPルーティング：github.com/labstack/echo/v4 v4.13.4
- データベース：database/sql（標準ライブラリ）
- MySQLドライバ：github.com/go-sql-driver/mysql v1.9.2

## プロジェクト構成

```
named-lock/
├── cmd/
│   ├── client/
│   │   └── main.go            # クライアントのメインエントリーポイント
│   └── server/
│       └── main.go            # サーバーのエントリーポイント
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
│   ├── post/
│   │   ├── common.go          # クライアント共通処理
│   │   ├── test_client.go     # テスト用クライアント
│   │   └── test_client_hold_release.go # ホールド・リリーステスト用クライアント
│   └── service/
│       └── lock_service.go    # ビジネスロジック
├── docker-compose.yml         # Docker Compose設定
├── go.mod                     # Goモジュール定義
├── go.sum                     # Goモジュール依存関係
└── README.md                  # このファイル
```

## セットアップと実行方法

### 1. MySQLコンテナの起動

```bash
docker compose up -d
```

### 2. APIサーバーの起動

```bash
go run cmd/server/main.go
```

### 3. テストクライアントの実行

別のターミナルを開いて、以下のコマンドを実行します。引数にクライアントIDと並列数を指定できます。

```bash
# メインクライアントの実行（通常のロックテスト）
go run cmd/client/main.go

# メインクライアントの実行（ロック保持・解放テスト）
go run cmd/client/main.go hold 10  # 10秒間ロックを保持

# 開始ID、並列数を指定して実行（例: ID 1から始まる5つのクライアント）
go run cmd/client/main.go 1 5

# 開始ID、並列数、テストモードを指定して実行
go run cmd/client/main.go 1 5 hold 10  # ID 1から5つのクライアントで10秒間ロック保持テスト
```

並列実行の場合、第1引数は開始クライアントID、第2引数は並列数を指定します。各クライアントは独自のIDを持ち、並行してロックの取得・解放を試みます。

テストモードには以下のオプションがあります：
- `normal`または`n`：通常のロック取得・解放テスト（デフォルト）
- `hold`または`h`：ロック保持・解放テスト（追加パラメータで保持時間を秒単位で指定可能）
- `process`または`p`：プロセスロックテスト（データ処理を含むロック取得・解放テスト）

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


### ロック取得・保持・解放（一連の操作）

```
POST /api/locks/hold-and-release
```

リクエスト例:
```json
{
  "lock_name": "test_lock",
  "timeout": 10,
  "hold_duration": 5
}
```

レスポンス例:
```json
{
  "success": true,
  "session_id": "123456",
  "message": "Lock acquired, held for 5 seconds, and released successfully. Current connection ID: 123456"
}
```

### ロック取得・処理・解放（一連の操作）

```
POST /api/locks/process
```

リクエスト例:
```json
{
  "lock_name": "test_process_lock",
  "item_key": "item1",
  "value": "{\"data\": \"updated_value\"}",
  "timeout": 10
}
```

レスポンス例:
```json
{
  "success": true,
  "session_id": "123456",
  "message": "Process completed successfully for item: item1"
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

-- ロックを解放
SELECT RELEASE_LOCK('test_lock');
```
