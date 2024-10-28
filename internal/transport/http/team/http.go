package team

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/team"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints team.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service user.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Get("/teams/{id}", kithttp.NewServer(endpoints.GetTeam, decodeGetTeamRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/teams", kithttp.NewServer(endpoints.GetTeams, decodeGetTeamsRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/teams/cities/{city_id}", kithttp.NewServer(endpoints.GetTeamsByCity, decodeGetTeamsByCityRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/teams/leagues/{league_id}", kithttp.NewServer(endpoints.GetTeamsByLeague, decodeGetTeamsByLeagueRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// r.Get("/teams/vs/{city_id}", kithttp.NewServer(endpoints.GetTeamVsTeamTable, decodeGetTeamVsTeamTableRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	//todo: Авторизация для Create
	r.Post("/teams", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Put("/teams", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Delete("/teams/{id}", kithttp.NewServer(endpoints.Delete, decodeDeleteRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.Post("/teams/add-player", kithttp.NewServer(endpoints.AddPlayerIntoTeam, decodeAddPlayerIntoTeamRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Post("/teams/remove-player", kithttp.NewServer(endpoints.RemovePlayerFromTeam, decodeRemovePlayerFromTeamRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
