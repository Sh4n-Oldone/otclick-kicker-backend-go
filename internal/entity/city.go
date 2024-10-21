package entity

type City struct{
	ID int64	`db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Ru string `db:"ru" json:"ru"`
	Deleted bool `db:"deleted" json:"deleted"`
}

type GetCityListRequest struct{
	WithDeleted bool
}

type CreateCityRequest struct{
	Name string `db:"name" json:"name"`
	Ru string `db:"ru" json:"ru"`
	Deleted bool `db:"deleted" json:"deleted,omitempty"`
}

type UpdateCityRequest struct{
	ID int64	`db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Ru string `db:"ru" json:"ru"`
	Deleted bool `db:"deleted" json:"deleted"`
}

type DeleteCityRequest struct{
	ID int64
}

type GetCityListResponse struct {
	Cities []City	`json:"cities"`
}