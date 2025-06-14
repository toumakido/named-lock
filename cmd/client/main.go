package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/example/named-lock/internal/post"
)

func main() {
	// コマンドライン引数の解析
	startID, parallelCount, args := post.ParseCommonArgs()

	// テストモードの選択（デフォルトは通常のロックテスト）
	testMode := "normal"
	if len(args) > 0 {
		testMode = args[0]
	}

	// テストモードに応じて処理を分岐
	switch testMode {
	case "normal", "n":
		fmt.Println("実行モード: 通常のロック取得・解放テスト")
		post.RunNormalLockTest(startID, parallelCount)
	case "hold", "h":
		fmt.Println("実行モード: ロック保持・解放テスト")
		// 保持時間の取得（デフォルト: 5秒）
		holdDuration := 5
		if len(args) > 1 {
			if duration, err := strconv.Atoi(args[1]); err == nil && duration > 0 {
				holdDuration = duration
			}
		}
		fmt.Printf("保持時間: %d秒\n", holdDuration)
		post.RunHoldReleaseLockTest(startID, parallelCount, holdDuration)
	default:
		fmt.Printf("未知のテストモード: %s\n", testMode)
		fmt.Println("使用方法: go run ./client [開始ID] [並列数] [テストモード] [追加パラメータ...]")
		fmt.Println("テストモード:")
		fmt.Println("  normal, n: 通常のロック取得・解放テスト")
		fmt.Println("  hold, h: ロック保持・解放テスト [保持時間(秒)]")
		os.Exit(1)
	}
}
