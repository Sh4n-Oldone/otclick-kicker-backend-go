package entities

type CreateSeasonRequest struct {
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type Season struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type CreateSeasonResponse struct {
	Id int64 `db:"id" json:"id"`
}

type UpdateSeasonRequest struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type DeleteSeasonRequest struct {
	ID int64 `json:"id"`
}

type GetSeasonResponse struct {
	Season []Season `json:"season"`
}
