package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	customerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func ValidateCreateCityRequest(request *entities.CreateCityRequest) error {
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

func ValidateUpdateCityRequest(request *entities.UpdateCityRequest) error {
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

func ValidateCreateUserRequest(request *entities.CreateUserRequest) error {
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

func ValidateCreateMatchRequest(request *entities.CreateMatchRequest) error {
	if request.Date.IsZero() {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Date))
	}

	return nil
}

func ValidateUpdateMatchRequest(request *entities.UpdateMatchRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateDeleteTeamRequest(request *entities.DeleteTeamRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	return nil
}

func ValidateGetTeamRequest(request *entities.GetTeamRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}
	return nil
}

func ValidateChangePasswordRequest(request *entities.ChangePasswordRequest) error {
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

func ValidateCreateTableRequest(request *entities.CreateTableRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func ValidateUpdateTableRequest(request *entities.UpdateTableRequest) error {
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

func ValidateDeleteTableRequest(request *entities.DeleteTableRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreateBarRequest(request *entities.CreateBarRequest) error {
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

func ValidateUpdateBarRequest(request *entities.UpdateBarRequest) error {
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

func ValidateDeleteBarRequest(request *entities.DeleteBarRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreatePlaceRequest(request *entities.CreatePlaceRequest) error {
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

func ValidateUpdatePlaceRequest(request *entities.UpdatePlaceRequest) error {
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

func ValidateDeletePlaceRequest(request *entities.DeletePlaceRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateCreateLeagueRequest(request *entities.CreateLeagueRequest) error {
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

func ValidateUpdateLeagueRequest(request *entities.UpdateLeagueRequest) error {
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

func ValidateCreateExtraPointsRequest(request *entities.CreateExtraPointsRequest) error {
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

func ValidateUpdateExtraPointsRequest(request *entities.UpdateExtraPointsRequest) error {
	if request.Id <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Id)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func ValidateTeamLeagueIdRequest(request *entities.TeamLeagueIdRequest) error {
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

func ValidateIdRequest(request *entities.IdRequest) error {
	if request.Id <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Id)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	return nil
}

func ValidateCreateSeasonRequest(request *entities.CreateSeasonRequest) error {
	if request.Name == "" {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.Name))
	}

	return nil
}

func ValidateUpdateSeasonRequest(request *entities.UpdateSeasonRequest) error {
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

func ValidateDeleteSeasonRequest(request *entities.DeleteSeasonRequest) error {
	if request.ID <= 0 {
		err := error_templates.New("invalid request fields", errors.New("invalid request fields"), codes.InvalidArgument, http.StatusBadRequest)
		return error_templates.WrapErrorDetail(err, fmt.Sprintf("wrong value of parameter %T", request.ID))
	}

	return nil
}

func ValidateDeleteFutureGame(request entities.DeleteFutureGameRequest) error {
	if request.Team1ID <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Team1ID)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	if request.Team2ID <= 0 {
		err := fmt.Errorf("wrong value of parameter %T", request.Team2ID)
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	return nil
}

func ValidateGetGameList(request entities.GetGameListRequest) error {

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(request)
	if err != nil {
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	return nil
}

func ValidCreateTournamentMasterRequest(request *entities.CreateTournamentMasterRequest, v *validator.Validate) error {
	err := v.Struct(request)
	if err != nil {
		return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}
	return nil
}
