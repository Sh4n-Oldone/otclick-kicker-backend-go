package tournament

import (
	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/tournament"
)

type Endpoints struct {
	GetTournamentTypeList  endpoint.Endpoint
	Create                 endpoint.Endpoint
	Update                 endpoint.Endpoint
	Delete                 endpoint.Endpoint
	FinishStage            endpoint.Endpoint
	GetTournamentStageList endpoint.Endpoint
	GetTournamentList      endpoint.Endpoint
}

func MakeEndpoints(s tournament.IService) Endpoints {
	return Endpoints{
		GetTournamentTypeList:  makeGetTournamentTypeList(s),
		Create:                 makeCreate(s),
		Update:                 makeUpdate(s),
		Delete:                 makeDelete(s),
		FinishStage:            makeFinishStage(s),
		GetTournamentStageList: makeGetTournamentStageList(s),
		GetTournamentList:      makeGetTournamentList(s),
	}
}
