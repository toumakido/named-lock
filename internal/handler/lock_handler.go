package handler

import (
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

// LockResponse はロック操作レスポンスの構造体
type LockResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// AcquireLock はロックを取得するハンドラ
func (h *LockHandler) AcquireLock(c echo.Context) error {
	var req AcquireLockRequest
	if err := c.Bind(&req); err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Failed to get current session ID: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// ロックを取得
	success, sessionID, err := h.lockService.AcquireLock(req.LockName, req.Timeout)
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success:   false,
			SessionID: "",
			Message:   "Failed to acquire lock: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// レスポンスを作成
	response := LockResponse{
		Success:   success,
		SessionID: sessionID,
	}

	if success {
		response.Message = "Lock acquired successfully. Current connection ID: " + currentSessionID
	} else {
		// ロックが取得できなかった場合、所有者を確認
		hasOwner, ownerID, err := h.lockService.GetLockOwner(req.LockName)
		if err != nil {
			// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
			response := LockResponse{
				Success:   false,
				SessionID: sessionID,
				Message:   "Failed to get lock owner: " + err.Error(),
			}
			return c.JSON(http.StatusOK, response)
		}

		if hasOwner {
			response.Message = "Lock is already held by session ID: " + ownerID
		} else {
			response.Message = "Failed to acquire lock"
		}
	}

	return c.JSON(http.StatusOK, response)
}

// ReleaseLock はロックを解放するハンドラ
func (h *LockHandler) ReleaseLock(c echo.Context) error {
	lockName := c.Param("lockName")

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success: false,
			Message: "Failed to get current session ID: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// ロックを解放
	success, err := h.lockService.ReleaseLock(lockName)
	if err != nil {
		// エラーが発生した場合でも、LockResponse形式でレスポンスを返す
		response := LockResponse{
			Success:   false,
			SessionID: currentSessionID,
			Message:   "Failed to release lock: " + err.Error(),
		}
		return c.JSON(http.StatusOK, response)
	}

	// レスポンスを作成
	response := LockResponse{
		Success:   success,
		SessionID: currentSessionID,
	}

	if success {
		response.Message = "Lock released successfully"
	} else {
		response.Message = "Failed to release lock. It may be held by another session or not exist."
	}

	return c.JSON(http.StatusOK, response)
}

// GetLockStatus はロックの状態を取得するハンドラ
func (h *LockHandler) GetLockStatus(c echo.Context) error {
	lockName := c.Param("lockName")

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		// エラーが発生した場合でも、適切な形式でレスポンスを返す
		return c.JSON(http.StatusOK, map[string]interface{}{
			"lock_name": lockName,
			"is_locked": false,
			"error":     "Failed to get current session ID: " + err.Error(),
			"success":   false,
		})
	}

	// ロックの所有者を取得
	hasOwner, ownerID, err := h.lockService.GetLockOwner(lockName)
	if err != nil {
		// エラーが発生した場合でも、適切な形式でレスポンスを返す
		return c.JSON(http.StatusOK, map[string]interface{}{
			"lock_name":          lockName,
			"is_locked":          false,
			"current_session_id": currentSessionID,
			"error":              "Failed to get lock owner: " + err.Error(),
			"success":            false,
		})
	}

	// ロックが解放されているかを確認（情報としてログに出力するなど必要に応じて使用）
	_, err = h.lockService.IsLockFree(lockName)
	if err != nil {
		// エラーが発生した場合でも、適切な形式でレスポンスを返す
		return c.JSON(http.StatusOK, map[string]interface{}{
			"lock_name":          lockName,
			"is_locked":          hasOwner,
			"owner_session_id":   ownerID,
			"current_session_id": currentSessionID,
			"error":              "Failed to check if lock is free: " + err.Error(),
			"success":            false,
		})
	}

	// レスポンスを作成
	type LockStatusResponse struct {
		LockName                string `json:"lock_name"`
		IsLocked                bool   `json:"is_locked"`
		OwnerSessionID          string `json:"owner_session_id,omitempty"`
		CurrentSessionID        string `json:"current_session_id"`
		IsOwnedByCurrentSession bool   `json:"is_owned_by_current_session"`
	}

	response := LockStatusResponse{
		LockName:         lockName,
		IsLocked:         hasOwner,
		CurrentSessionID: currentSessionID,
	}

	if hasOwner {
		response.OwnerSessionID = ownerID
		response.IsOwnedByCurrentSession = ownerID == currentSessionID
	}

	return c.JSON(http.StatusOK, response)
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

// RegisterRoutes はルートを登録する
func (h *LockHandler) RegisterRoutes(e *echo.Echo) {
	e.POST("/api/locks", h.AcquireLock)
	e.DELETE("/api/locks/:lockName", h.ReleaseLock)
	e.GET("/api/locks/:lockName", h.GetLockStatus)
	e.GET("/api/session", h.GetCurrentSession)
	e.POST("/api/locks/hold-and-release", h.AcquireHoldReleaseLock)
}
