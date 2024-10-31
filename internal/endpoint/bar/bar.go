package bar

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/bar"
)

type Endpoints struct {
	GetList endpoint.Endpoint
	Create endpoint.Endpoint
	Update endpoint.Endpoint
	Delete endpoint.Endpoint
}

func MakeEndpoints(s bar.IService) Endpoints {
	return Endpoints{
		GetList: makeGetList(s),
		Create: makeCreate(s),
		Update: makeUpdate(s),
		Delete: makeDelete(s),
	}
}
