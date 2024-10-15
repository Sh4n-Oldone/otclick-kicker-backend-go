package helpers

import (
	"errors"
	"fmt"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"net/http"

	"node71.otclick.ru/backend/template/pkg/error_templates"
)

func CastValidateRequest[T proto.Message](validator *protovalidate.Validator, request any) (T, error) {
	req, ok := request.(T)
	if !ok {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return *new(T), error_templates.WrapErrorDetail(err, fmt.Sprintf("cannot cast request to %T", req))
	}

	err := validator.Validate(req)
	if err != nil {
		return req, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return req, nil
}
