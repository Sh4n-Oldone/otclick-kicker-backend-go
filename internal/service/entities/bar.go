package entities

import "time"

type Bar struct {
	ID          int64 `db:"id" json:"id"`
	City        City
	Name        string     `db:"name" json:"name"`
	Description string     `db:"description" json:"description"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

type BarShort struct {
	ID   *int64  `db:"id" json:"id"`
	Name *string `db:"name" json:"name"`
}

type GetBarListRequest struct {
	CityID      *int64
	WithDeleted bool
}

type GetBarListResponse struct {
	Bars []Bar `json:"bars"`
}

type CreateBarRequest struct {
	CityID      int64  `json:"cityId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateBarRequest struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type DeleteBarRequest struct {
	ID int64
}
