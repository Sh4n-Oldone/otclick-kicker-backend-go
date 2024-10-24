package entity

type Team struct{
	ID int64 `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	ShortName string `json:"short_name" db:"short_name"`
	Avatar string `json:"avatar" db:"avatar"`
	League League
	City City
}

