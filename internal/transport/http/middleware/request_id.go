package middleware

import (
	"context"
	"net/http"
	"strings"

	guuid "github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const RequestIDKey = "X-Request-ID"

func GetRequestID(ctx context.Context) (reqID string, newCtx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	if val, ok := md[strings.ToLower(RequestIDKey)]; ok {
		if val[0] != "" {
			reqID = val[0]
		}
	}
	if reqID == "" {
		reqID = guuid.New().String()
	}
	newMd := metadata.New(map[string]string{RequestIDKey: reqID})
	newCtx = metadata.NewOutgoingContext(ctx, newMd)
	return reqID, newCtx
}

func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDKey)
		md := metadata.Pairs(RequestIDKey, requestID)
		ctx = metadata.NewIncomingContext(ctx, md)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}
