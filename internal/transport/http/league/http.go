package league

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/league"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints league.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service user.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Get("/leagues", kithttp.NewServer(endpoints.GetList, decodeGetListRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Post("/leagues", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Put("/leagues/{id}", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Delete("/leagues/{id}", kithttp.NewServer(endpoints.Delete, decodeDeleteRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Post("/leagues/{id}/recalc", kithttp.NewServer(endpoints.Recalc, decodeRecalcRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Post("/extra-points", kithttp.NewServer(endpoints.CreateExtraPoints, decodeCreateExtraPointsRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Patch("/extra-points", kithttp.NewServer(endpoints.UpdateExtraPoints, decodeUpdateExtraPointsRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Delete("/extra-points/{id}", kithttp.NewServer(endpoints.DeleteExtraPoints, decodeDeleteExtraPointsRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/leagues/{league_id}/teams/{team_id}/extra-points", kithttp.NewServer(endpoints.GetExtraPointsListByTeamAndLeagueId, decodeGetExtraPointsListByTeamAndLeagueIdRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/extra-points/{id}", kithttp.NewServer(endpoints.GetExtraPointsById, decodeGetExtraPointsByIdRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
