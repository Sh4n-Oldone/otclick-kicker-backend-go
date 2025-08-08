package player

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/player"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

func NewServer(endpoints player.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service user.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Get("/players/find", kithttp.NewServer(endpoints.FindPlayers, decodeFindPlayersRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/players/{id}", kithttp.NewServer(endpoints.Get, decodeGetRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/players/team/{teamId}", kithttp.NewServer(endpoints.GetByTeam, decodeGetByTeamIDRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminTournamentMaster()).
		Post("/players", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Delete("/players/{id}", kithttp.NewServer(endpoints.Delete, decodeDeleteRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Post("/players/{id}/recover", kithttp.NewServer(endpoints.Recover, decodeRecoverRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Patch("/players/{id}", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
