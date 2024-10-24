package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"

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

func ValidateCreateUserRequest(request *entity.CreateUserRequest) error {
	if request.Email == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Email))
	}
	_, _err := mail.ParseAddress(request.Email)
	if _err != nil {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Email))
	}
	if request.Password == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Password))
	}
	if len(request.Password) < 6 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Password))
	}

	return nil
}

func ValidateCreateMatchRequest(request *entity.CreateMatchRequest) error {
	if request.Date == nil {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Date))
	}

	return nil
}

func ValidateUpdateMatchRequest(request *entity.UpdateMatchRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateChangePasswordRequest(request *entity.ChangePasswordRequest) error {
	if request.Email == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Email))
	}
	_, _err := mail.ParseAddress(request.Email)
	if _err != nil {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Email))
	}
	if request.Password == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Password))
	}
	if len(request.Password) < 6 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("too short value of parameter %T", request.Password))
	}
	if request.NewPassword == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.NewPassword))
	}
	if len(request.NewPassword) < 6 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("too short value of parameter %T", request.NewPassword))
	}

	return nil
}

