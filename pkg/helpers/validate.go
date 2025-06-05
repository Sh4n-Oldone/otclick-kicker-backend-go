package helpers

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"

	"net/http"
	"net/mail"
	"time"

	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	customerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
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
	if request.Date.IsZero() {
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

func ValidateCreateTeamRequest(request *entity.CreateTeamRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}
	if request.ShortName == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ShortName))
	}
	return nil
}

func ValidateUpdateTeamRequest(request *entity.UpdateTeamRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	return nil
}

func ValidateDeleteTeamRequest(request *entity.DeleteTeamRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	return nil
}

func ValidatePlayerTeamRequest(request *entity.PlayerTeam) error {
	if request.PlayerID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.PlayerID))
	}
	if request.TeamID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.TeamID))
	}
	return nil
}

func ValidateGetTeamRequest(request *entity.GetTeamRequest) error {
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

func ValidateCreateTableRequest(request *entity.CreateTableRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func ValidateUpdateTableRequest(request *entity.UpdateTableRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func ValidateDeleteTableRequest(request *entity.DeleteTableRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreateBarRequest(request *entity.CreateBarRequest) error {
	if request.CityID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.CityID))
	}
	if request.CityID == 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.CityID))
	}

	return nil
}

func ValidateUpdateBarRequest(request *entity.UpdateBarRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	if request.Name != nil {
		if *request.Name == "" {
			err := errors.New(customerr.ErrEmptyField)
			err = error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
		}
	}
	if request.Description != nil {
		if *request.Description == "" {
			err := errors.New(customerr.ErrEmptyField)
			err = error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Description))
		}
	}

	return nil
}

func ValidateDeleteBarRequest(request *entity.DeleteBarRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreatePlaceRequest(request *entity.CreatePlaceRequest) error {
	if request.BarID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.BarID))
	}
	if request.TableID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.TableID))
	}

	return nil
}

func ValidateUpdatePlaceRequest(request *entity.UpdatePlaceRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	if request.BarID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.BarID))
	}
	if request.TableID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.TableID))
	}

	return nil
}

func ValidateDeletePlaceRequest(request *entity.DeletePlaceRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreateLeagueRequest(request *entity.CreateLeagueRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}
	if request.CityID == 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.CityID))
	}

	return nil
}

func ValidateUpdateLeagueRequest(request *entity.UpdateLeagueRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func NewCustomValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())

	_ = validate.RegisterValidation("valid-date", validateDate)

	return validate
}

func validateDate(fl validator.FieldLevel) bool {
	date := fl.Field().Interface().(time.Time)
	return !(date.Equal(time.Unix(0, 0)) || date.IsZero() || date.Year() == 0)
}

func ValidateCreateExtraPointsRequest(request *entity.CreateExtraPointsRequest) error {
	if request.TeamId <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.TeamId)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	if request.LeagueId <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.LeagueId)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	if request.Reason == "" {
		err := fmt.Errorf("wrong value of parameter %T", request.Reason)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func ValidateUpdateExtraPointsRequest(request *entity.UpdateExtraPointsRequest) error {
	if request.Id <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Id)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func ValidateTeamLeagueIdRequest(request *entity.TeamLeagueIdRequest) error {
	if request.TeamId <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.TeamId)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	if request.LeagueId <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.LeagueId)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	return nil
}

func ValidateIdRequest(request *entity.IdRequest) error {
	if request.Id <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Id)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func ValidateCreateSeasonRequest(request *entity.CreateSeasonRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func ValidateUpdateSeasonRequest(request *entity.UpdateSeasonRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	if request.Name != nil {
		if *request.Name == "" {
			err := errors.New(customerr.ErrEmptyField)
			err = error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
		}
	}
	if request.Description != nil {
		if *request.Description == "" {
			err := errors.New(customerr.ErrEmptyField)
			err = error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
			return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Description))
		}
	}

	return nil
}

func ValidateDeleteSeasonRequest(request *entity.DeleteSeasonRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}
