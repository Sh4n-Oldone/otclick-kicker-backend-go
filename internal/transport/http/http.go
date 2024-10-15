package http

import (
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

func RunHTTPServer(httpServer *http.Server, l net.Listener, log zerolog.Logger, listenErr chan error) {
	log.Info().Msgf("starting http server on %s", l.Addr())
	if err := httpServer.Serve(l); err != nil && err != http.ErrServerClosed {
		listenErr <- err
	}
}
