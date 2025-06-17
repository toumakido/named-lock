package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

// OrderLockRequest は注文ロックリクエストの構造体
type OrderLockRequest struct {
	ProductCode string `json:"product_code"`
	Timeout     int    `json:"timeout"`
}

// OrderLockResponse は注文ロックレスポンスの構造体
type OrderLockResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// AcquireOrderReleaseLock はロックを取得し、注文処理し、解放する
func (c *Client) AcquireOrderReleaseLock(productCode string, timeout int) (*OrderLockResponse, error) {
	reqBody, err := json.Marshal(OrderLockRequest{
		ProductCode: productCode,
		Timeout:     timeout,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post("http://localhost:8080/api/locks/order", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var orderResp OrderLockResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

// RunOrderTest は注文ロックテストを実行する
func RunOrderTest(c *Client, productCode string, args ...interface{}) {
	// 実行開始時間を記録
	startTime := time.Now()

	// ロックを取得・処理・解放
	fmt.Printf("Client %d [%.1fs]: Acquiring, processing, and releasing lock for order with product code: %s\n",
		c.ID, time.Since(startTime).Seconds(), productCode)
	orderResp, err := c.AcquireOrderReleaseLock(productCode, -1)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Operation failed: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Operation result: %+v\n", c.ID, time.Since(startTime).Seconds(), orderResp)
}

// RunOrderLockTest は注文ロックテストを実行する関数
func RunOrderLockTest(startID int, parallelCount int) {
	productCode := uuid.New().String() // ランダムな商品コードを生成
	RunParallel(startID, parallelCount, productCode, RunOrderTest)
}
