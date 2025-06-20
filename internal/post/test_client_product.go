package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// ProductLockRequest はプロセスロックリクエストの構造体
type ProductLockRequest struct {
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	Timeout     int    `json:"timeout"`
}

// ProductLockResponse はプロセスロックレスポンスの構造体
type ProductLockResponse struct {
	Success   bool                   `json:"success"`
	SessionID string                 `json:"session_id,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

// AcquireProductReleaseLock はロックを取得し、処理し、解放する
func (c *Client) AcquireProductReleaseLock(productCode string, quantity int, timeout int) (*ProductLockResponse, error) {
	reqBody, err := json.Marshal(ProductLockRequest{
		ProductCode: productCode,
		Quantity:    quantity,
		Timeout:     timeout,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post("http://localhost:8080/api/locks/product", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var processResp ProductLockResponse
	if err := json.Unmarshal(body, &processResp); err != nil {
		return nil, err
	}

	return &processResp, nil
}

// RunProductTest はプロセスロックテストを実行する
func RunProductTest(c *Client, productCode string, args ...interface{}) {
	// 実行開始時間を記録
	startTime := time.Now()

	// 数量を設定（デフォルトは1）
	quantity := 1
	if len(args) > 0 {
		if q, ok := args[0].(int); ok {
			quantity = q
		}
	}

	// ロックを取得・処理・解放
	fmt.Printf("Client %d [%.1fs]: Acquiring, processing, and releasing lock for product: %s, quantity: %d\n",
		c.ID, time.Since(startTime).Seconds(), productCode, quantity)
	processResp, err := c.AcquireProductReleaseLock(productCode, quantity, -1)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Operation failed: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Operation result: %+v\n", c.ID, time.Since(startTime).Seconds(), processResp)
}

// RunProductLockTest はプロセスロックテストを実行する関数
func RunProductLockTest(startID int, parallelCount int) {
	productCode := uuid.New().String() // ランダムな商品コードを生成
	RunParallel(startID, parallelCount, productCode, RunProductTest)
}
