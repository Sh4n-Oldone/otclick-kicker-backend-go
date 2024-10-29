package game

import (
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/game"
)

type Endpoints struct {
	Create           endpoint.Endpoint
	Delete           endpoint.Endpoint
	Get              endpoint.Endpoint
	Update           endpoint.Endpoint
	Find             endpoint.Endpoint
	UpdateFutureGame endpoint.Endpoint
	GetYears         endpoint.Endpoint
	GetComingGames   endpoint.Endpoint
	GetFutureGames   endpoint.Endpoint
	CreateFutureGame endpoint.Endpoint
	GetTeamGames     endpoint.Endpoint
}

func MakeEndpoints(s game.IService) Endpoints {
	return Endpoints{
		Create:           makeCreate(s),
		Delete:           makeDelete(s),
		Get:              makeGet(s),
		Update:           makeUpdate(s),
		Find:             makeFind(s),
		UpdateFutureGame: makeUpdateFutureGame(s),
		GetYears:         makeGetYears(s),
		GetComingGames:   makeGetComingGames(s),
		GetFutureGames:   makeGetFutureGames(s),
		CreateFutureGame: makeCreateFutureGame(s),
		GetTeamGames:     makeGetTeamGames(s),
	}
}
