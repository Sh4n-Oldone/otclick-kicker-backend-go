package team

import (
	"github.com/go-kit/kit/endpoint"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/team"
)

type Endpoints struct {
	GetTeam          endpoint.Endpoint
	GetTeams         endpoint.Endpoint
	GetTeamsByCity   endpoint.Endpoint
	GetTeamsByLeague endpoint.Endpoint
	// GetTeamVsTeamTable   endpoint.Endpoint
	Create               endpoint.Endpoint
	Update               endpoint.Endpoint
	Delete               endpoint.Endpoint
	AddPlayerIntoTeam    endpoint.Endpoint
	RemovePlayerFromTeam endpoint.Endpoint
}

func MakeEndpoints(s team.IService) Endpoints {
	return Endpoints{
		GetTeam:          makeGetTeam(s),
		GetTeams:         makeGetTeams(s),
		GetTeamsByCity:   makeGetTeamsByCity(s),
		GetTeamsByLeague: makeGetTeamsByLeague(s),
		// GetTeamVsTeamTable:   makeGetTeamVsTeamTable(s),
		Create:               makeCreate(s),
		Update:               makeUpdate(s),
		Delete:               makeDelete(s),
		AddPlayerIntoTeam:    makeAddPlayerIntoTeam(s),
		RemovePlayerFromTeam: makeRemovePlayerFromTeam(s),
	}
}
