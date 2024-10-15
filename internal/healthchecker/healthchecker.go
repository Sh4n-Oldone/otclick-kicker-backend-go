package healthchecker

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type HealthChecker struct {
	healthChecker healthcheck.Handler
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		healthChecker: healthcheck.NewMetricsHandler(prometheus.DefaultRegisterer, "health_check"),
	}
}

func (hc *HealthChecker) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.NoCache)
	r.Get("/_live", hc.LiveEndpointHTTP())
	r.Get("/_ready", hc.ReadyEndpointHTTP())
	return r
}

func (hc *HealthChecker) GetHealthChecker() healthcheck.Handler {
	return hc.healthChecker
}

func (hc *HealthChecker) LiveEndpointHTTP() http.HandlerFunc {
	return hc.healthChecker.LiveEndpoint
}

func (hc *HealthChecker) ReadyEndpointHTTP() http.HandlerFunc {
	return hc.healthChecker.ReadyEndpoint

}

func (hc *HealthChecker) LiveEndpointFastHTTP() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hc.healthChecker.LiveEndpoint(w, r)
	}))
}

func (hc *HealthChecker) ReadyEndpointFastHTTP() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hc.healthChecker.ReadyEndpoint(w, r)
	}))
}
