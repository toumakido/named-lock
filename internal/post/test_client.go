package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

// 通常のロック取得・解放テスト
func RunNormalTest(c *Client, lockName string, args ...interface{}) {
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

// 通常のロックテストを実行する関数
func RunNormalLockTest(startID int, parallelCount int) {
	lockName := "test_lock"
	RunParallel(startID, parallelCount, lockName, RunNormalTest)
}
