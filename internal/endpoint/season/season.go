package season

import (
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/season"
)

type Endpoints struct {
	GetList endpoint.Endpoint
	Create  endpoint.Endpoint
	Update  endpoint.Endpoint
	Delete  endpoint.Endpoint
}

func MakeEndpoints(s season.IService) Endpoints {
	return Endpoints{
		GetList: makeGetList(s),
		Create:  makeCreate(s),
		Update:  makeUpdate(s),
		Delete:  makeDelete(s),
	}
}
