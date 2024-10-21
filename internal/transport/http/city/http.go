package city

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	kithttp "github.com/go-kit/kit/transport/http"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/endpoint/city"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/common"
	custom_middleware "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
)

// NewServer initializes a new http server
func NewServer(endpoints city.Endpoints, options []kithttp.ServerOption) http.Handler {
	r := chi.NewRouter()

	options = append(options, kithttp.ServerErrorEncoder(common.EncodeErrorResponse))

	r.Use(middleware.NoCache)
	r.Use(middleware.RealIP)
	r.Use(custom_middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(custom_middleware.HeaderHandler)

	//Actual
	r.Get("/cities", kithttp.NewServer(endpoints.GetList, decodeGetListRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Post("/cities", kithttp.NewServer(endpoints.Create, decodeCreateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Put("/cities", kithttp.NewServer(endpoints.Update, decodeUpdateRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)
	r.Delete("/cities/delete/{id}", kithttp.NewServer(endpoints.Delete, decodeDeleteRequest, kithttp.EncodeJSONResponse, options...).ServeHTTP)


	return r
}
