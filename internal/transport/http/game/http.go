package game

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/game"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints game.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service user.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Get("/games/played/{id}", kithttp.NewServer(endpoints.Get, decodeIdParamRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Get("/games/future", kithttp.NewServer(endpoints.GetFutureGames, decodeGetFutureGamesRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// /games/find-games - поиск игр
	r.Get("/games/find-games", kithttp.NewServer(endpoints.Find, decodeFindRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// /games/years возвращает все годы, в которых есть хоть одна игра
	r.Get("/games/years", kithttp.NewServer(endpoints.GetYears, decodeGamesYearsRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// /games/coming возвращает список ближайших игр
	r.Get("/games/coming", kithttp.NewServer(endpoints.GetComingGames, decodeGetComingGamesRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// /games возвращает список игр (есть фильтр)
	r.Get("/games", kithttp.NewServer(endpoints.GetGameList, decodeGetGameListRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Post("/games/played", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Post("/games/future", kithttp.NewServer(endpoints.CreateFutureGame, decodeCreateFutureGameRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdmin()).
		Delete("/games/played/{id}", kithttp.NewServer(endpoints.Delete, decodeIdParamRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// Patch "/games/played" - обновляет игру с матчами, пересчитывает рейтинг.
	// Рейтинг считается правильно только если переданная игра - последняя, в которой участвовали эти игроки
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Patch("/games/played", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Patch("/games/future", kithttp.NewServer(endpoints.UpdateFutureGame, decodeUpdateFutureGameRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	// Get "/games/teams/{team_id}" - получает список будущих домашних игр для указанной команды
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Get("/games/teams/{team_id}", kithttp.NewServer(endpoints.GetTeamGames, decodeGetTeamIDRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthSuperUserAdminCapitan()).
		Delete("/games/future/{id}", kithttp.NewServer(endpoints.DeleteFutureGame, decodeDeleteFutureGameRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	return r
}
