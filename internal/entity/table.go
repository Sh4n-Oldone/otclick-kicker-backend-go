package entity

import "time"

type Table struct{
	ID int64	`db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type GetTableListRequest struct{
	WithDeleted bool
}

type GetTableListResponse struct{
	Tables []Table	
}

type GetTableRequest struct{
	ID int64
}

type GetTableResponse struct{
	Table Table	
}

type CreateTableRequest struct{
	Name string `json:"name"`
}

type UpdateTableRequest struct{
	ID int64	`json:"id"`
	Name string `json:"name"`
}

type DeleteTableRequest struct{
	ID int64
}
