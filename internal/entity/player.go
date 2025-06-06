package entity

type FullPlayer struct {
	ID           int     `json:"id"`
	Name         *string `json:"name,omitempty"`
	SecondName   *string `json:"secondName,omitempty"`
	LastName     string  `json:"lastName"`
	Avatar       []byte  `json:"avatar,omitempty"`
	ActivePlayer *bool   `json:"activePlayer,omitempty"`
	Deleted      bool    `json:"deleted"`
	CityID       *int    `json:"cityId,omitempty"`
	CityName     *string `json:"cityName,omitempty"`

	Leagues []LeagueItem `json:"leagues,omitempty"`
}
