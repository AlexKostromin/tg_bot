// internal/api/dto/user.go

package dto

type UpdateUserRequest struct {
	IsActive bool `json:"is_active"`
}

type UserResponse struct {
	ID           int    `json:"id"`
	ChatID       int64  `json:"chat_id"`
	FullName     string `json:"full_name"`
	Phone        string `json:"phone"`
	ClassNumber  int    `json:"class_number"`
	ClassGroupID int    `json:"class_group_id"`
	GroupName    string `json:"group_name"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

type UserListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}
