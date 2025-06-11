package post

import (
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
		Client: &http.Client{Timeout: 30 * time.Minute}, // タイムアウトを30分に設定
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

// 共通の引数解析関数
func ParseCommonArgs() (startID int, parallelCount int, args []string) {
	startID = 1
	parallelCount = 1

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

	// 残りの引数を返す
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}

	return
}

// 並列実行のためのヘルパー関数
func RunParallel(startID int, parallelCount int, lockName string, runFunc func(client *Client, lockName string, args ...interface{}), args ...interface{}) {
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
			runFunc(client, lockName, args...)
		}(clientID)
	}

	// すべてのクライアントの完了を待機
	wg.Wait()
	fmt.Println("All clients completed")
}
