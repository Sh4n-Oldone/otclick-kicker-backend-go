package middleware

import (
	"google.golang.org/grpc/metadata"
	"net/http"
)

func HeaderHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		mdCTX := metadata.NewIncomingContext(ctx, metadata.MD(r.Header))
		next.ServeHTTP(w, r.WithContext(mdCTX))
	}
	return http.HandlerFunc(fn)
}
