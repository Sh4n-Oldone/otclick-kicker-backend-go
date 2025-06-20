package entities

import "time"

type Place struct {
	ID        int64 `db:"id" json:"id"`
	Bar       Bar
	Table     Table
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" db:"deleted_at"`
}

type PlaceShort struct {
	ID    *int64     `db:"id" json:"id"`
	Bar   BarShort   `db:"bar" json:"bar"`
	Table TableShort `db:"table" json:"table"`
}

type GetPlaceListRequest struct {
	BarID       *int64
	TableID     *int64
	CityID      *int64
	WithDeleted bool
}

type GetPlaceListResponse struct {
	Places []Place
}

type GetPlaceRequest struct {
	ID int64
}

type GetPlaceResponse struct {
	Place Place
}

type CreatePlaceRequest struct {
	BarID   int64 `json:"barId"`
	TableID int64 `json:"tableId"`
}

type UpdatePlaceRequest struct {
	ID      int64 `json:"id"`
	BarID   int64 `json:"barId"`
	TableID int64 `json:"tableId"`
}

type DeletePlaceRequest struct {
	ID int64
}
