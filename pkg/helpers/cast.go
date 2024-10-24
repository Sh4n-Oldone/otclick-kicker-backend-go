package helpers

import (
	stderr "errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"net/http"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func CastRequest[T any](request any) (T, error) {
	req, ok := request.(T)
	if !ok {
		err := error_templates.New("invalid request fields", stderr.New(errors.FailedCastRequest), codes.InvalidArgument, http.StatusBadRequest)
		return *new(T), error_templates.WrapErrorDetail(err, fmt.Sprintf("cannot cast request to %T", req))
	}
	return req, nil
}
