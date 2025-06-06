package entity

import "time"

type Role struct {
	ID          int64      `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" db:"updated_at"`
}

type GetRoleListResponse struct {
	Roles []Role
}
