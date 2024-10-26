package entity

import "time"

type Place struct{
	ID           int64  	`db:"id" json:"id"`
	Bar Bar
	Table Table
	UpdatedAt 	 time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt 	*time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
}

type GetPlaceListRequest struct{
	BarID      *int64
	TableID    *int64
	WithDeleted bool
}

type GetPlaceListResponse struct{
	Places []Place	
}

type GetPlaceRequest struct{
	ID int64
}

type GetPlaceResponse struct{
	Place Place
}

type CreatePlaceRequest struct{
	BarID           int64  	`json:"bar_id"`
	TableID           int64  	`json:"table_id"`
}

type UpdatePlaceRequest struct{
	ID int64	`json:"id"`
	BarID           int64  	`json:"bar_id"`
	TableID           int64  	`json:"table_id"`
}

type DeletePlaceRequest struct{
	ID int64
}
