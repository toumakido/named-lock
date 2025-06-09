package handler

import (
	"encoding/json"
	"net/http"

	"github.com/example/named-lock/internal/service"
	"github.com/gorilla/mux"
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

// LockResponse はロック操作レスポンスの構造体
type LockResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// AcquireLock はロックを取得するハンドラ
func (h *LockHandler) AcquireLock(w http.ResponseWriter, r *http.Request) {
	var req AcquireLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		http.Error(w, "Failed to get current session ID", http.StatusInternalServerError)
		return
	}

	// ロックを取得
	success, sessionID, err := h.lockService.AcquireLock(req.LockName, req.Timeout)
	if err != nil {
		http.Error(w, "Failed to acquire lock: "+err.Error(), http.StatusInternalServerError)
		return
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
			http.Error(w, "Failed to get lock owner: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if hasOwner {
			response.Message = "Lock is already held by session ID: " + ownerID
		} else {
			response.Message = "Failed to acquire lock"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ReleaseLock はロックを解放するハンドラ
func (h *LockHandler) ReleaseLock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lockName := vars["lockName"]

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		http.Error(w, "Failed to get current session ID", http.StatusInternalServerError)
		return
	}

	// ロックを解放
	success, err := h.lockService.ReleaseLock(lockName)
	if err != nil {
		http.Error(w, "Failed to release lock: "+err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLockStatus はロックの状態を取得するハンドラ
func (h *LockHandler) GetLockStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lockName := vars["lockName"]

	// 現在のセッションIDを取得
	currentSessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		http.Error(w, "Failed to get current session ID", http.StatusInternalServerError)
		return
	}

	// ロックの所有者を取得
	hasOwner, ownerID, err := h.lockService.GetLockOwner(lockName)
	if err != nil {
		http.Error(w, "Failed to get lock owner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ロックが解放されているかを確認（情報としてログに出力するなど必要に応じて使用）
	_, err = h.lockService.IsLockFree(lockName)
	if err != nil {
		http.Error(w, "Failed to check if lock is free: "+err.Error(), http.StatusInternalServerError)
		return
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentSession は現在のセッションIDを取得するハンドラ
func (h *LockHandler) GetCurrentSession(w http.ResponseWriter, r *http.Request) {
	// 現在のセッションIDを取得
	sessionID, err := h.lockService.GetCurrentSessionID()
	if err != nil {
		http.Error(w, "Failed to get current session ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンスを作成
	type SessionResponse struct {
		SessionID string `json:"session_id"`
	}

	response := SessionResponse{
		SessionID: sessionID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes はルートを登録する
func (h *LockHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/locks", h.AcquireLock).Methods("POST")
	router.HandleFunc("/api/locks/{lockName}", h.ReleaseLock).Methods("DELETE")
	router.HandleFunc("/api/locks/{lockName}", h.GetLockStatus).Methods("GET")
	router.HandleFunc("/api/session", h.GetCurrentSession).Methods("GET")
}
