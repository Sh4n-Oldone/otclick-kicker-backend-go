package user

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
)

type Endpoints struct {
	Create         endpoint.Endpoint
	Login          endpoint.Endpoint
	ChangePassword endpoint.Endpoint
	CheckAuth	   endpoint.Endpoint
}

func MakeEndpoints(s user.IService) Endpoints {
	return Endpoints{
		Create:         makeCreate(s),
		Login:          makeLogin(s),
		ChangePassword: makeChangePassword(s),
		CheckAuth: 		makeCheckAuth(s),
	}
}
