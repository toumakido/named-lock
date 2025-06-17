package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ProcessLockRequest はプロセスロックリクエストの構造体
type ProcessLockRequest struct {
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	Timeout     int    `json:"timeout"`
}

// ProcessLockResponse はプロセスロックレスポンスの構造体
type ProcessLockResponse struct {
	Success   bool                   `json:"success"`
	SessionID string                 `json:"session_id,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

// AcquireProcessReleaseLock はロックを取得し、処理し、解放する
func (c *Client) AcquireProcessReleaseLock(productCode string, quantity int, timeout int) (*ProcessLockResponse, error) {
	reqBody, err := json.Marshal(ProcessLockRequest{
		ProductCode: productCode,
		Quantity:    quantity,
		Timeout:     timeout,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post("http://localhost:8080/api/locks/process", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var processResp ProcessLockResponse
	if err := json.Unmarshal(body, &processResp); err != nil {
		return nil, err
	}

	return &processResp, nil
}

// RunProcessTest はプロセスロックテストを実行する
func RunProcessTest(c *Client, productCode string, args ...interface{}) {
	// 実行開始時間を記録
	startTime := time.Now()

	// 数量を設定（デフォルトは1）
	quantity := 1
	if len(args) > 0 {
		if q, ok := args[0].(int); ok {
			quantity = q
		}
	}

	// ロックの状態を確認
	status, err := c.GetLockStatus(productCode)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock status before operation: %+v\n", c.ID, time.Since(startTime).Seconds(), status)

	// ロックを取得・処理・解放
	fmt.Printf("Client %d [%.1fs]: Acquiring, processing, and releasing lock for product: %s, quantity: %d\n",
		c.ID, time.Since(startTime).Seconds(), productCode, quantity)
	processResp, err := c.AcquireProcessReleaseLock(productCode, quantity, -1)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Operation failed: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Operation result: %+v\n", c.ID, time.Since(startTime).Seconds(), processResp)

	// ロックの状態を最終確認
	status, err = c.GetLockStatus(productCode)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock status after operation: %+v\n", c.ID, time.Since(startTime).Seconds(), status)
}

// RunProcessLockTest はプロセスロックテストを実行する関数
func RunProcessLockTest(startID int, parallelCount int) {
	productCode := "P001" // テスト用の商品コード
	RunParallel(startID, parallelCount, productCode, RunProcessTest)
}
