package endpoint

import (
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/role"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/match"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/player"
)

type ServicesEndpoints struct {
	CityEP city.Endpoints
	UserEP user.Endpoints
	RoleEP role.Endpoints
	MatchEP match.Endpoints
	PlayerEP player.Endpoints
}
