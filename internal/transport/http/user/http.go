package user

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/config"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/user"
	userService "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints user.Endpoints, options []kithttp.ServerOption, cfg *config.Configuration, service userService.IService) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Post("/login", kithttp.NewServer(endpoints.Login, decodeLoginRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service), custom_middleware.AuthNotCaptain()).
		Post("/users", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	r.With(custom_middleware.Auth(cfg, service)).
		Put("/users/change-password", kithttp.NewServer(endpoints.ChangePassword, decodeChangePasswordRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
