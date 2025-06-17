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

// Product は商品情報を表す構造体
type Product struct {
	ID    int
	Code  string
	Name  string
	Price float64
}

// Inventory は在庫情報を表す構造体
type Inventory struct {
	ID        int
	ProductID int
	Quantity  int
}

// Order は注文情報を表す構造体
type Order struct {
	ID        int
	ProductID int
	Quantity  int
	Status    string
}

// GetProductByCode は商品コードから商品情報を取得する
func (tx *Tx) GetProductByCode(productCode string) (*Product, error) {
	var product Product

	query := `
		SELECT id, product_code, name, price 
		FROM products 
		WHERE product_code = ?`

	err := tx.QueryRow(query, productCode).Scan(&product.ID, &product.Code, &product.Name, &product.Price)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 商品が存在しない場合はnilを返す
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetInventoryForUpdate は商品IDから在庫情報をFOR UPDATE句を使用して取得する
func (tx *Tx) GetInventoryForUpdate(productID int) (*Inventory, error) {
	var inventory Inventory

	query := `
		SELECT id, product_id, quantity 
		FROM product_inventory 
		WHERE product_id = ? 
		FOR UPDATE`

	err := tx.QueryRow(query, productID).Scan(&inventory.ID, &inventory.ProductID, &inventory.Quantity)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 在庫情報が存在しない場合はnilを返す
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return &inventory, nil
}

// GetOrderForUpdate は商品IDから注文情報をFOR UPDATE句を使用して取得する
func (tx *Tx) GetOrderForUpdate(productID int) (*Order, error) {
	var order Order

	query := `
		SELECT id, product_id, quantity, status 
		FROM orders 
		WHERE product_id = ? AND status = 'pending'
		ORDER BY id ASC
		LIMIT 1
		FOR UPDATE`

	err := tx.QueryRow(query, productID).Scan(&order.ID, &order.ProductID, &order.Quantity, &order.Status)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // 注文情報が存在しない場合はnilを返す
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// UpdateInventory は在庫情報を更新する
func (tx *Tx) UpdateInventory(inventory *Inventory) error {
	query := `
		UPDATE product_inventory 
		SET quantity = ?
		WHERE id = ?`

	_, err := tx.Exec(query, inventory.Quantity, inventory.ID)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// InsertInventory は新しい在庫情報を挿入する
func (tx *Tx) InsertInventory(inventory *Inventory) error {
	query := `
		INSERT INTO product_inventory 
		(product_id, quantity) 
		VALUES (?, ?)`

	result, err := tx.Exec(query, inventory.ProductID, inventory.Quantity)
	if err != nil {
		return fmt.Errorf("failed to insert inventory: %w", err)
	}

	// 挿入されたIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	inventory.ID = int(id)

	return nil
}

// UpdateOrder は注文情報を更新する
func (tx *Tx) UpdateOrder(order *Order) error {
	query := `
		UPDATE orders 
		SET quantity = ?, status = ?
		WHERE id = ?`

	_, err := tx.Exec(query, order.Quantity, order.Status, order.ID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return nil
}

// InsertOrder は新しい注文情報を挿入する
func (tx *Tx) InsertOrder(order *Order) error {
	query := `
		INSERT INTO orders 
		(product_id, quantity, status) 
		VALUES (?, ?, ?)`

	result, err := tx.Exec(query, order.ProductID, order.Quantity, order.Status)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// 挿入されたIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	order.ID = int(id)

	return nil
}
