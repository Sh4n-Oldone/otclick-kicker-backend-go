package player

import (
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/player"
)

type Endpoints struct {
	Create                  endpoint.Endpoint
	Delete                  endpoint.Endpoint
	Recover                 endpoint.Endpoint
	Update                  endpoint.Endpoint
	Get                     endpoint.Endpoint
	GetByTeam               endpoint.Endpoint
	FindPlayers             endpoint.Endpoint
	GetTournamentPlayerList endpoint.Endpoint
}

func MakeEndpoints(s player.IService) Endpoints {
	return Endpoints{
		Create:                  makeCreate(s),
		Delete:                  makeDelete(s),
		Recover:                 makeRecover(s),
		Update:                  makeUpdate(s),
		Get:                     makeGet(s),
		GetByTeam:               makeGetByTeamID(s),
		FindPlayers:             makeFindPlayers(s),
		GetTournamentPlayerList: makeGetTournamentPlayerList(s),
	}
}
