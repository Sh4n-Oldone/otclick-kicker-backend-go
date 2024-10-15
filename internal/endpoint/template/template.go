package template

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/backend/template/internal/service/template"
)

type Endpoints struct {
	Create endpoint.Endpoint
}

func MakeEndpoints(s template.IService) Endpoints {
	return Endpoints{
		Create: makeCreate(s),
	}
}
