package entity

import "time"

type City struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Ru        string     `db:"ru" json:"ru"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
}

type GetCityListRequest struct {
	WithDeleted bool
}

type CreateCityRequest struct {
	Name      string     `db:"name" json:"name"`
	Ru        string     `db:"ru" json:"ru"`
	DeletedAt *time.Time `json:"deletedAt" db:"deleted_at"`
}

type UpdateCityRequest struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Ru   string `db:"ru" json:"ru"`
}

type DeleteCityRequest struct {
	ID int64
}

type GetCityListResponse struct {
	Cities []City `json:"cities"`
}
