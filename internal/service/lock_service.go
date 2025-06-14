package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/example/named-lock/internal/db"
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

// AcquireLock はロックを取得する
func (s *LockService) AcquireLock(lockName string, timeout int) (bool, string, error) {
	// 現在のセッションIDを取得
	sessionID, err := s.db.GetCurrentConnectionID()
	if err != nil {
		return false, "", fmt.Errorf("failed to get connection id: %w", err)
	}

	// ロックを取得
	result, err := s.db.GetNamedLock(lockName, timeout)
	if err != nil {
		return false, "", fmt.Errorf("failed to acquire lock: %w", err)
	}

	// ロック取得に成功した場合
	if result == 1 {
		// ロック履歴を保存（オプション）
		sessionIDStr := strconv.FormatInt(sessionID, 10)
		if err := s.db.SaveLockHistory(lockName, sessionIDStr, "acquired"); err != nil {
			// 履歴保存に失敗してもロック自体は取得できているので、エラーはログに出すだけ
			fmt.Printf("failed to save lock history: %v\n", err)
		}
		return true, fmt.Sprintf("%d", sessionID), nil
	}

	return false, "", nil
}

// ReleaseLock はロックを解放する
func (s *LockService) ReleaseLock(lockName string) (bool, error) {
	// 現在のセッションIDを取得
	sessionID, err := s.db.GetCurrentConnectionID()
	if err != nil {
		return false, fmt.Errorf("failed to get connection id: %w", err)
	}

	// ロックを解放
	result, err := s.db.ReleaseNamedLock(lockName)
	if err != nil {
		return false, fmt.Errorf("failed to release lock: %w", err)
	}

	// ロック解放に成功した場合
	if result == 1 {
		// ロック履歴を更新（オプション）
		sessionIDStr := strconv.FormatInt(sessionID, 10)
		if err := s.db.SaveLockHistory(lockName, sessionIDStr, "released"); err != nil {
			// 履歴更新に失敗してもロック自体は解放できているので、エラーはログに出すだけ
			fmt.Printf("failed to update lock history: %v\n", err)
		}
		return true, nil
	}

	return false, nil
}

// GetLockOwner はロックの所有者を取得する
func (s *LockService) GetLockOwner(lockName string) (bool, string, error) {
	// ロックの所有者を取得
	result, err := s.db.GetLockOwner(lockName)
	if err != nil {
		return false, "", fmt.Errorf("failed to get lock owner: %w", err)
	}

	// ロックが取得されている場合
	if result.Valid {
		return true, fmt.Sprintf("%d", result.Int64), nil
	}

	return false, "", nil
}

// IsLockFree はロックが解放されているかを確認する
func (s *LockService) IsLockFree(lockName string) (bool, error) {
	// ロックが解放されているかを確認
	result, err := s.db.IsFreeLock(lockName)
	if err != nil {
		return false, fmt.Errorf("failed to check if lock is free: %w", err)
	}

	// ロックが存在しない場合
	if !result.Valid {
		return true, nil
	}

	// ロックが解放されている場合
	return result.Int64 == 1, nil
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
func (s *LockService) AcquireHoldReleaseLock(lockName string, timeout int, holdDuration int) (bool, string, error) {
	// ロックを取得
	acquired, sessionID, err := s.AcquireLock(lockName, timeout)
	if err != nil {
		return false, "", fmt.Errorf("failed to acquire lock: %w", err)
	}

	// ロック取得に失敗した場合
	if !acquired {
		return false, "", nil
	}

	// 指定された時間だけ待機
	time.Sleep(time.Duration(holdDuration) * time.Second)

	// ロックを解放
	released, err := s.ReleaseLock(lockName)
	if err != nil {
		return true, sessionID, fmt.Errorf("acquired lock but failed to release: %w", err)
	}

	if !released {
		return true, sessionID, fmt.Errorf("acquired lock but failed to release: unknown error")
	}

	return true, sessionID, nil
}
