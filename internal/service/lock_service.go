package service

import (
	"context"
	"fmt"
	"time"

	"github.com/example/named-lock/internal/db"
	"github.com/google/uuid"
	"github.com/samber/do"
)

// LockService はロック操作に関するサービス
type LockService struct {
	db *db.DB
}

// NewLockService は新しいLockServiceインスタンスを作成する
func NewLockService(injector *do.Injector) (*LockService, error) {
	database := do.MustInvoke[*db.DB](injector)
	return &LockService{
		db: database,
	}, nil
}

// GetCurrentSessionID は現在のセッションIDを取得する
func (s *LockService) GetCurrentSessionID() (string, error) {
	sessionID, err := s.db.GetCurrentConnectionID()
	if err != nil {
		return "", fmt.Errorf("failed to get connection id: %w", err)
	}
	return fmt.Sprintf("%d", sessionID), nil
}

// AcquireHoldReleaseLock はロックを取得し、指定された時間保持した後、解放する
// ロックの取得と解放の間にトランザクションを張る
func (s *LockService) AcquireHoldReleaseLock(ctx context.Context, lockName string, timeout int, holdDuration int) (string, error) {
	id := uuid.New().String()

	// トランザクションを開始
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sID, err := tx.GetCurrentConnectionID()
	if err != nil {
		return "", fmt.Errorf("failed to get connection id: %w", err)
	}
	fmt.Printf("[%s]before lock session ID:%d", id, sID)
	sessionID := fmt.Sprintf("%d", sID)

	// ロックを取得
	result, err := tx.GetNamedLock(lockName, timeout)
	if err != nil {
		return sessionID, fmt.Errorf("failed to acquire lock: %w", err)
	}
	// ロック取得に失敗した場合
	if result != 1 {
		return sessionID, fmt.Errorf("failed to acquire lock: result %d", result)
	}

	sID, err = tx.GetCurrentConnectionID()
	if err != nil {
		return "", fmt.Errorf("failed to get connection id: %w", err)
	}
	fmt.Printf("[%s]after lock session ID:%d", id, sID)

	// 指定された時間だけ待機
	time.Sleep(time.Duration(holdDuration) * time.Second)

	sID, err = tx.GetCurrentConnectionID()
	if err != nil {
		return "", fmt.Errorf("failed to get connection id: %w", err)
	}
	fmt.Printf("[%s]before release session ID:%d", id, sID)

	// ロックを解放
	result, err = tx.ReleaseNamedLock(lockName)
	if err != nil {
		return sessionID, fmt.Errorf("failed to release: %w", err)
	}
	if result != 1 {
		return sessionID, fmt.Errorf("failed to release: result %d", result)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return sessionID, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sessionID, nil
}
