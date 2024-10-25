package entity

type Year struct {
	Year uint16 `json:"year"`
}

type GetGamesYearsResponse struct {
	Years []Year `json:"years"`
}
