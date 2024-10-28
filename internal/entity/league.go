package entity

type GetLeagueListRequest struct {
	CityID int64 `db:"city_id" json:"cityId"`
}

type GetLeagueListResponse struct {
	Leagues []League `json:"leagues"`
}

type League struct {
	ID     int64  `db:"id" json:"id"`
	Name   string `db:"name" json:"name"`
	CityID int64  `db:"city_id" json:"cityId"`
}

type CreateLeagueRequest struct {
	CityID int64   `db:"city_id" json:"cityId"`
	Name   string  `db:"name" json:"name"`
	Teams  []int64 `json:"teams"`
}

type CreateLeagueResponse struct {
	ID int64 `db:"id" json:"id"`
}

type UpdateLeagueRequest struct {
	ID    int64   `db:"id" json:"id"`
	Name  string  `db:"name" json:"name"`
	Teams []int64 `json:"teams"`
}

type UpdateLeagueResponse struct {
	ID int64 `db:"id" json:"id"`
}

type DeleteLeagueRequest struct {
	ID int64
}
