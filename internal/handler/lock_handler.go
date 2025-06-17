package handler

import (
	"fmt"
	"net/http"

	"github.com/example/named-lock/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

// LockHandler はロック操作に関するHTTPハンドラ
type LockHandler struct {
	lockService *service.LockService
}

// NewLockHandler は新しいLockHandlerインスタンスを作成する
func NewLockHandler(injector *do.Injector) (*LockHandler, error) {
	lockService := do.MustInvoke[*service.LockService](injector)
	return &LockHandler{
		lockService: lockService,
	}, nil
}

// AcquireLockRequest はロック取得リクエストの構造体
type AcquireLockRequest struct {
	LockName string `json:"lock_name"`
	Timeout  int    `json:"timeout"`
}

// AcquireHoldReleaseRequest はロック取得・保持・解放リクエストの構造体
type AcquireHoldReleaseRequest struct {
	LockName     string `json:"lock_name"`
	Timeout      int    `json:"timeout"`
	HoldDuration int    `json:"hold_duration"`
}

// AcquireProcessReleaseRequest はロック取得・処理・解放リクエストの構造体
type AcquireProcessReleaseRequest struct {
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	Timeout     int    `json:"timeout"`
}

// LockResponse はロック操作レスポンスの構造体
type LockResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// GetCurrentSession は現在のセッションIDを取得するハンドラ
func (h *LockHandler) GetCurrentSession(c echo.Context) error {
	// 現在のセッションIDを取得
	sessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		// エラーが発生した場合でも、適切な形式でレスポンスを返す
		return c.JSON(http.StatusOK, map[string]interface{}{
			"error":   "Failed to get current session ID: " + err.Error(),
			"success": false,
		})
	}

	// レスポンスを作成
	type SessionResponse struct {
		SessionID string `json:"session_id"`
	}

	response := SessionResponse{
		SessionID: sessionID,
	}

	return c.JSON(http.StatusOK, response)
}

// AcquireHoldReleaseLock はロックを取得し、指定された時間保持した後、解放するハンドラ
func (h *LockHandler) AcquireHoldReleaseLock(c echo.Context) error {
	var req AcquireHoldReleaseRequest
	if err := c.Bind(&req); err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// ロックを取得し、保持し、解放する
	sessionID, err := h.lockService.AcquireHoldReleaseLock(c.Request().Context(), req.LockName, req.Timeout, req.HoldDuration)
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success:   false,
			SessionID: sessionID, // エラー時でもセッションIDがある場合は返す
			Message:   "Operation failed: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}
	success := true

	// レスポンスを作成
	response := LockResponse{
		Success:   success,
		SessionID: sessionID,
	}

	return c.JSON(http.StatusOK, response)
}

// AcquireProcessReleaseLock はロックを取得し、処理し、解放するハンドラ
func (h *LockHandler) AcquireProcessReleaseLock(c echo.Context) error {
	var req AcquireProcessReleaseRequest
	if err := c.Bind(&req); err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// ロックを取得し、処理し、解放する
	err := h.lockService.AcquireProcessReleaseLock(c.Request().Context(), req.ProductCode, req.Quantity, req.Timeout)
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Operation failed: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// セッションIDを取得
	sessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Failed to get session ID: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// レスポンスを作成
	response := LockResponse{
		Success:   true,
		SessionID: sessionID,
		Message:   fmt.Sprintf("Process completed successfully for product: %s, quantity: %d", req.ProductCode, req.Quantity),
	}

	return c.JSON(http.StatusOK, response)
}

// RegisterRoutes はルートを登録する
func (h *LockHandler) RegisterRoutes(e *echo.Echo) {
	e.GET("/api/session", h.GetCurrentSession)
	e.POST("/api/locks/hold-and-release", h.AcquireHoldReleaseLock)
	e.POST("/api/locks/process", h.AcquireProcessReleaseLock)
}
