package post

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ロックを取得し、保持し、解放する
func (c *Client) AcquireHoldReleaseLock(lockName string, timeout int, holdDuration int) (*LockResponse, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"lock_name":     lockName,
		"timeout":       timeout,
		"hold_duration": holdDuration,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post("http://localhost:8080/api/locks/hold-and-release", "application/json", bytes.NewBuffer(reqBody))
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

// ホールド＆リリーステスト
func RunHoldReleaseTest(c *Client, lockName string, args ...interface{}) {
	// 保持時間を取得
	holdDuration := 5 // デフォルト値
	if len(args) > 0 {
		if duration, ok := args[0].(int); ok {
			holdDuration = duration
		}
	}
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
	// status, err := c.GetLockStatus(lockName)
	// if err != nil {
	// 	fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
	// 	return
	// }
	// fmt.Printf("Client %d [%.1fs]: Lock status before operation: %+v\n", c.ID, time.Since(startTime).Seconds(), status)

	// ロックを取得・保持・解放
	fmt.Printf("Client %d [%.1fs]: Acquiring, holding for %d seconds, and releasing lock...\n", c.ID, time.Since(startTime).Seconds(), holdDuration)
	lockResp, err := c.AcquireHoldReleaseLock(lockName, -1, holdDuration)
	if err != nil {
		fmt.Printf("Client %d [%.1fs]: Operation failed: %v\n", c.ID, time.Since(startTime).Seconds(), err)
		return
	}
	fmt.Printf("Client %d [%.1fs]: Operation result: %+v\n", c.ID, time.Since(startTime).Seconds(), lockResp)

	// ロックの状態を最終確認
	// status, err = c.GetLockStatus(lockName)
	// if err != nil {
	// 	fmt.Printf("Client %d [%.1fs]: Failed to get lock status: %v\n", c.ID, time.Since(startTime).Seconds(), err)
	// 	return
	// }
	// fmt.Printf("Client %d [%.1fs]: Lock status after operation: %+v\n", c.ID, time.Since(startTime).Seconds(), status)
}

// ホールド＆リリーステストを実行する関数
func RunHoldReleaseLockTest(startID int, parallelCount int, holdDuration int) {
	lockName := "test_lock"
	RunParallel(startID, parallelCount, lockName, RunHoldReleaseTest, holdDuration)
}
