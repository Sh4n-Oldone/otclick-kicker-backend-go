package role

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/role"
)

type Endpoints struct {
	GetList endpoint.Endpoint
}

func MakeEndpoints(s role.IService) Endpoints {
	return Endpoints{
		GetList: makeGetList(s),
	}
}
