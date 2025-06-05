package league

import (
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/league"
)

type Endpoints struct {
	GetList                             endpoint.Endpoint
	Create                              endpoint.Endpoint
	Update                              endpoint.Endpoint
	Delete                              endpoint.Endpoint
	Recalc                              endpoint.Endpoint
	CreateExtraPoints                   endpoint.Endpoint
	UpdateExtraPoints                   endpoint.Endpoint
	DeleteExtraPoints                   endpoint.Endpoint
	GetExtraPointsListByTeamAndLeagueId endpoint.Endpoint
	GetExtraPointsById                  endpoint.Endpoint
}

func MakeEndpoints(s league.IService) Endpoints {
	return Endpoints{
		GetList:                             makeGetList(s),
		Create:                              makeCreate(s),
		Update:                              makeUpdate(s),
		Delete:                              makeDelete(s),
		Recalc:                              makeRecalc(s),
		CreateExtraPoints:                   makeCreateExtraPoints(s),
		UpdateExtraPoints:                   makeUpdateExtraPoints(s),
		DeleteExtraPoints:                   makeDeleteExtraPoints(s),
		GetExtraPointsListByTeamAndLeagueId: makeGetExtraPointsListByTeamAndLeagueId(s),
		GetExtraPointsById:                  makeGetExtraPointsById(s),
	}
}
