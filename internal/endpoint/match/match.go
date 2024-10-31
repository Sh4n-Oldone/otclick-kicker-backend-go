package match

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/match"
)

type Endpoints struct {
	Create endpoint.Endpoint
	Update endpoint.Endpoint
	Delete endpoint.Endpoint
}

func MakeEndpoints(s match.IService) Endpoints {
	return Endpoints{
		Create: makeCreate(s),
		Update: makeUpdate(s),
		Delete: makeDelete(s),
	}
}
