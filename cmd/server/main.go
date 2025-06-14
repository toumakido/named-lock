package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/named-lock/internal/config"
	"github.com/example/named-lock/internal/db"
	"github.com/example/named-lock/internal/handler"
	"github.com/example/named-lock/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/samber/do"
)

func main() {
	// 依存性注入コンテナを作成
	injector := do.New()

	// 設定を登録
	cfg := config.NewConfig()
	do.Provide(injector, func(i *do.Injector) (*config.Config, error) {
		return cfg, nil
	})

	// データベース接続を登録
	do.Provide(injector, func(i *do.Injector) (*db.DB, error) {
		dbCfg := do.MustInvoke[*config.Config](i).DB
		return db.NewDB(&dbCfg)
	})

	// サービスを登録
	do.Provide(injector, service.NewLockService)

	// ハンドラを登録
	do.Provide(injector, handler.NewLockHandler)

	// Echoインスタンスを作成
	e := echo.New()

	// ミドルウェアを設定
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// ハンドラを取得してルートを登録
	lockHandler := do.MustInvoke[*handler.LockHandler](injector)
	lockHandler.RegisterRoutes(e)

	// サーバーを起動
	go func() {
		log.Printf("Server is running on http://localhost:8080")
		if err := e.Start(":8080"); err != nil {
			log.Printf("Shutting down the server: %v", err)
		}
	}()

	// シグナルを待機
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// サーバーを停止
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	// データベース接続を閉じる
	database := do.MustInvoke[*db.DB](injector)
	if err := database.Close(); err != nil {
		log.Printf("Error closing database connection: %v\n", err)
	}

	log.Println("Server exiting")
}
