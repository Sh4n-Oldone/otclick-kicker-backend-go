package entity

type League struct{
	ID int64 `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	City City
}
