package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/example/named-lock/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

// DB はデータベース操作を行うための構造体
type DB struct {
	*sql.DB
}

// Tx はトランザクションを表す構造体
type Tx struct {
	*sql.Tx
}

// NewDB は新しいDBインスタンスを作成する
func NewDB(cfg *config.DBConfig) (*DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続テスト
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database successfully")
	return &DB{DB: db}, nil
}

// Close はデータベース接続を閉じる
func (db *DB) Close() error {
	return db.DB.Close()
}

// GetNamedLock は名前付きロックを取得する
// lockName: ロック名
// timeout: タイムアウト（秒）
// 戻り値: 1=ロック取得成功, 0=ロック取得失敗, error=エラー
func (db *DB) GetNamedLock(lockName string, timeout int) (int, error) {
	var result sql.NullInt64
	err := db.QueryRow("SELECT GET_LOCK(?, ?)", lockName, timeout).Scan(&result)
	if err != nil {
		return -1, fmt.Errorf("failed to get lock: %w", err)
	}
	return int(result.Int64), nil
}

// ReleaseNamedLock は名前付きロックを解放する
// lockName: ロック名
// 戻り値: 1=ロック解放成功, 0=ロックが存在しないか他のセッションが所有, error=エラー
func (db *DB) ReleaseNamedLock(lockName string) (int, error) {
	var result sql.NullInt64
	err := db.QueryRow("SELECT RELEASE_LOCK(?)", lockName).Scan(&result)
	if err != nil {
		return -1, fmt.Errorf("failed to release lock: %w", err)
	}
	return int(result.Int64), nil
}

// IsFreeLock はロックが解放されているかを確認する
// lockName: ロック名
// 戻り値: 1=ロックは解放されている, 0=ロックは取得されている, NULL=ロックが存在しない, error=エラー
func (db *DB) IsFreeLock(lockName string) (sql.NullInt64, error) {
	var result sql.NullInt64
	err := db.QueryRow("SELECT IS_FREE_LOCK(?)", lockName).Scan(&result)
	if err != nil {
		return sql.NullInt64{}, fmt.Errorf("failed to check if lock is free: %w", err)
	}
	return result, nil
}

// GetLockOwner はロックを所有しているセッションIDを取得する
// lockName: ロック名
// 戻り値: セッションID（ロックが存在しない場合はNULL）, error=エラー
func (db *DB) GetLockOwner(lockName string) (sql.NullInt64, error) {
	var result sql.NullInt64
	err := db.QueryRow("SELECT IS_USED_LOCK(?)", lockName).Scan(&result)
	if err != nil {
		return sql.NullInt64{}, fmt.Errorf("failed to get lock owner: %w", err)
	}
	return result, nil
}

// GetCurrentConnectionID は現在の接続のセッションIDを取得する
func (db *DB) GetCurrentConnectionID() (int64, error) {
	var id int64
	err := db.QueryRow("SELECT CONNECTION_ID()").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection id: %w", err)
	}
	return id, nil
}

// SaveLockHistory はロック取得履歴を保存する（オプション機能）
func (db *DB) SaveLockHistory(lockName string, sessionID string, status string) error {
	var query string
	if status == "acquired" {
		query = "INSERT INTO lock_history (lock_name, session_id, status) VALUES (?, ?, ?)"
		_, err := db.Exec(query, lockName, sessionID, status)
		return err
	} else {
		query = "UPDATE lock_history SET released_at = CURRENT_TIMESTAMP, status = ? WHERE lock_name = ? AND session_id = ? AND status = 'acquired'"
		_, err := db.Exec(query, status, lockName, sessionID)
		return err
	}
}

// BeginTx はトランザクションを開始する
// トランザクションを開始しても同じ接続（セッション）が使用されるため、セッションIDは変わらない
func (db *DB) BeginTx(ctx context.Context) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{Tx: tx}, nil
}

// GetNamedLock は名前付きロックを取得する
// lockName: ロック名
// timeout: タイムアウト（秒）
// 戻り値: 1=ロック取得成功, 0=ロック取得失敗, error=エラー
func (tx *Tx) GetNamedLock(lockName string, timeout int) (int, error) {
	var result sql.NullInt64
	err := tx.QueryRow("SELECT GET_LOCK(?, ?)", lockName, timeout).Scan(&result)
	if err != nil {
		return -1, fmt.Errorf("failed to get lock: %w", err)
	}
	return int(result.Int64), nil
}

// ReleaseNamedLock は名前付きロックを解放する
// lockName: ロック名
// 戻り値: 1=ロック解放成功, 0=ロックが存在しないか他のセッションが所有, error=エラー
func (tx *Tx) ReleaseNamedLock(lockName string) (int, error) {
	var result sql.NullInt64
	err := tx.QueryRow("SELECT RELEASE_LOCK(?)", lockName).Scan(&result)
	if err != nil {
		return -1, fmt.Errorf("failed to release lock: %w", err)
	}
	return int(result.Int64), nil
}

// GetCurrentConnectionID は現在の接続のセッションIDを取得する
func (tx *Tx) GetCurrentConnectionID() (int64, error) {
	var id int64
	err := tx.QueryRow("SELECT CONNECTION_ID()").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection id: %w", err)
	}
	return id, nil
}
