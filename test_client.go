package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// APIレスポンスの構造体
type SessionResponse struct {
	SessionID string `json:"session_id"`
}

type LockResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

type LockStatusResponse struct {
	LockName                string `json:"lock_name"`
	IsLocked                bool   `json:"is_locked"`
	OwnerSessionID          string `json:"owner_session_id,omitempty"`
	CurrentSessionID        string `json:"current_session_id"`
	IsOwnedByCurrentSession bool   `json:"is_owned_by_current_session"`
}

// クライアント構造体
type Client struct {
	ID     int
	Client *http.Client
}

// 新しいクライアントを作成
func NewClient(id int) *Client {
	return &Client{
		ID:     id,
		Client: &http.Client{Timeout: 30 * time.Second}, // タイムアウトを30秒に延長
	}
}

// セッションIDを取得
func (c *Client) GetSessionID() (string, error) {
	resp, err := c.Client.Get("http://localhost:8080/api/session")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return "", err
	}

	return sessionResp.SessionID, nil
}

// ロックを取得（リトライ機能付き）
func (c *Client) AcquireLock(lockName string, timeout int) (*LockResponse, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"lock_name": lockName,
		"timeout":   timeout,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post("http://localhost:8080/api/locks", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lockResp LockResponse
	if err := json.Unmarshal(body, &lockResp); err != nil {
		return nil, err
	}

	return &lockResp, nil
}

// ロックを解放
func (c *Client) ReleaseLock(lockName string) (*LockResponse, error) {
	req, err := http.NewRequest("DELETE", "http://localhost:8080/api/locks/"+lockName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lockResp LockResponse
	if err := json.Unmarshal(body, &lockResp); err != nil {
		return nil, err
	}

	return &lockResp, nil
}

// ロックの状態を取得
func (c *Client) GetLockStatus(lockName string) (*LockStatusResponse, error) {
	resp, err := c.Client.Get("http://localhost:8080/api/locks/" + lockName)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var statusResp LockStatusResponse
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, err
	}

	return &statusResp, nil
}

// クライアントの実行
func (c *Client) Run(lockName string) {
	// 実行開始時間を記録
	startTime := time.Now()

	// セッションIDを取得
	sessionID, err := c.GetSessionID()
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get session ID: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Session ID: %s\n", c.ID, time.Since(startTime).Seconds(), sessionID)

	// ロックの状態を確認
	status, err := c.GetLockStatus(lockName)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock status before acquiring: %+v\n", c.ID, time.Since(startTime).Seconds(), status)

	// ロックを取得
	lockResp, err := c.AcquireLock(lockName, -1)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to acquire lock: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock acquisition result: %+v\n", c.ID, time.Since(startTime).Seconds(), lockResp)

	// ロックの状態を再確認
	status, err = c.GetLockStatus(lockName)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock status after acquiring: %+v\n", c.ID, time.Since(startTime).Seconds(), status)

	// ロックを保持する時間
	holdTime := 0.5
	if lockResp.Success {
		fmt.Printf("Client %d [%.1fs]: Holding lock for %v seconds...\n", c.ID, time.Since(startTime).Seconds(), holdTime)
		time.Sleep(time.Duration(holdTime) * time.Second)
	}

	// ロックを解放
	releaseResp, err := c.ReleaseLock(lockName)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to release lock: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock release result: %+v\n", c.ID, time.Since(startTime).Seconds(), releaseResp)

	// ロックの状態を最終確認
	status, err = c.GetLockStatus(lockName)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Lock status after releasing: %+v\n", c.ID, time.Since(startTime).Seconds(), status)
}

func main() {
	// コマンドライン引数からクライアントIDと並列数を取得
	startID := 1
	parallelCount := 1

	// 第1引数: 開始クライアントID
	if len(os.Args) > 1 {
		id, err := strconv.Atoi(os.Args[1])
		if err == nil {
			startID = id
		}
	}

	// 第2引数: 並列数
	if len(os.Args) > 2 {
		count, err := strconv.Atoi(os.Args[2])
		if err == nil && count > 0 {
			parallelCount = count
		}
	}

	// ロック名
	lockName := "test_lock"

	// 並列実行のための同期グループ
	var wg sync.WaitGroup

	fmt.Printf("Starting %d clients in parallel (IDs: %d-%d)\n",
		parallelCount, startID, startID+parallelCount-1)

	// 指定された並列数だけクライアントを作成して実行
	for i := 0; i < parallelCount; i++ {
		wg.Add(1)

		clientID := startID + i

		// goroutineで並列実行
		go func(id int) {
			defer wg.Done()
			client := NewClient(id)
			client.Run(lockName)
		}(clientID)
	}

	// すべてのクライアントの完了を待機
	wg.Wait()
	fmt.Println("All clients completed")
}
