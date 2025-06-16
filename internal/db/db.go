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

// GetCurrentConnectionID は現在の接続のセッションIDを取得する
func (db *DB) GetCurrentConnectionID() (int64, error) {
	var id int64
	err := db.QueryRow("SELECT CONNECTION_ID()").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection id: %w", err)
	}
	return id, nil
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
