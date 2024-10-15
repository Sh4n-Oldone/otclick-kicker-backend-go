package template

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"node71.otclick.ru/backend/template/internal/endpoint/template"
	"node71.otclick.ru/backend/template/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/backend/template/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints template.Endpoints, options []kithttp.ServerOption) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Post("/template", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)

	return r
}
