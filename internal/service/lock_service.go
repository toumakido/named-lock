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
	if !result {
		return sessionID, fmt.Errorf("failed to acquire lock: result %v", result)
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
	if !result {
		return sessionID, fmt.Errorf("failed to release: result %v", result)
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return sessionID, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return sessionID, nil
}

// AcquireProductReleaseLock はロックを取得し、在庫を更新後、解放する
// ロックの取得と解放の間にトランザクションを張る
// 商品在庫を増やす処理を行う
func (s *LockService) AcquireProductReleaseLock(ctx context.Context, productCode string, addQuantity int, timeout int) error {
	id := uuid.New().String()

	// トランザクションを開始
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ロックを取得
	result, err := tx.GetNamedLock(productCode, timeout)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	// ロック取得に失敗した場合
	if !result {
		return fmt.Errorf("failed to acquire lock: result %v", result)
	}

	// 在庫情報を取得（FOR UPDATE句を使用）
	product, err := tx.GetProductForUpdate(productCode)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// 在庫情報が存在する場合は更新、存在しない場合は挿入
	if product != nil {
		fmt.Printf("[%s] Found existing product ID: %s, current quantity: %d\n", id, product.Code, product.Quantity)

		// 在庫数を増やす
		product.Quantity += addQuantity
		if err := tx.UpdateInventory(product); err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		fmt.Printf("[%s] Updated product quantity to: %d\n", id, product.Quantity)
	} else {
		fmt.Printf("[%s] No existing product found, inserting new product...\n", id)

		// 新しい在庫情報を挿入
		newProduct := &db.Product{
			Code:     productCode,
			Quantity: addQuantity,
		}
		if err := tx.InsertInventory(newProduct); err != nil {
			return fmt.Errorf("failed to insert newProduct: %w", err)
		}

		fmt.Printf("[%s] Inserted new product with quantity: %d\n", id, addQuantity)
	}

	// １秒まつ
	time.Sleep(1 * time.Second)

	// ロックを解放
	result, err = tx.ReleaseNamedLock(productCode)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !result {
		return fmt.Errorf("failed to release lock: result %v", result)
	}

	fmt.Printf("[%s] Lock released\n", id)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("[%s] Transaction committed\n", id)

	return nil
}

// AcquireOrderReleaseLock はロックを取得し、注文を挿入後、共通コードで取得して解放する
// ロックの取得と解放の間にトランザクションを張る
func (s *LockService) AcquireOrderReleaseLock(ctx context.Context, code string, timeout int) error {
	id := uuid.New().String()

	// トランザクションを開始
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// ロックを取得
	result, err := tx.GetNamedLock(code, timeout)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	// ロック取得に失敗した場合
	if !result {
		return fmt.Errorf("failed to acquire lock: result %v", result)
	}

	newOrder := &db.Order{
		ID:   id,
		Code: code,
	}
	if err := tx.InsertOrder(newOrder); err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	orders, err := s.db.ListOrderByCode(code)
	if err != nil {
		return fmt.Errorf("failed to list orders: %w", err)
	}

	fmt.Printf("[%s] Inserted new order with ID: %s, Code: %s, Total Orders: %d\n", id, newOrder.ID, newOrder.Code, len(orders))

	// ロックを解放
	result, err = tx.ReleaseNamedLock(code)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	if !result {
		return fmt.Errorf("failed to release lock: result %v", result)
	}

	fmt.Printf("[%s] Lock released\n", id)

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("[%s] Transaction committed\n", id)

	return nil
}
