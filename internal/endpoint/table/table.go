package table

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/table"
)

type Endpoints struct {
	GetList      endpoint.Endpoint
	Create       endpoint.Endpoint
	Update       endpoint.Endpoint
	Delete       endpoint.Endpoint
	GetTableByID endpoint.Endpoint
}

func MakeEndpoints(s table.IService) Endpoints {
	return Endpoints{
		GetList:      makeGetList(s),
		Create:       makeCreate(s),
		Update:       makeUpdate(s),
		Delete:       makeDelete(s),
		GetTableByID: makeGetTableByID(s),
	}
}
