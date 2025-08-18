package dto

import "ssh-key-manager/internal/model"

// APIResponse는 표준 API 응답 구조체입니다.
type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    interface{}     `json:"data,omitempty"`
	Error   *model.APIError `json:"error,omitempty"`
}
