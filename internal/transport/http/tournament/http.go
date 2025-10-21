package tournament

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/tournament"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

func NewServer(endpoints tournament.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service user.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	r.Get("/tournaments/types", kithttp.NewServer(endpoints.GetTournamentTypeList, decodeGetTournamentTypeListRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/tournaments/{id}/stages", kithttp.NewServer(endpoints.GetTournamentStageList, decodeIdRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/tournaments", kithttp.NewServer(endpoints.GetTournamentList, decodeGetTournamentListRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminTournamentMaster()).
		Post("/tournaments", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminTournamentMaster()).
		Patch("/tournaments/{id}", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminTournamentMaster()).
		Delete("/tournaments/{id}", kithttp.NewServer(endpoints.Delete, decodeDeleteRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminTournamentMaster()).
		Post("/tournaments/stages/{id}", kithttp.NewServer(endpoints.FinishStage, decodeFinishStageRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
