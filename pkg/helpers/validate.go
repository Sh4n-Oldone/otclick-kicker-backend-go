package helpers

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

func ValidateCreateCityRequest(request *entity.CreateCityRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}
	if request.Ru == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Ru))
	}

	return nil
}

func ValidateUpdateCityRequest(request *entity.UpdateCityRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}
	if request.Ru == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Ru))
	}

	return nil
}