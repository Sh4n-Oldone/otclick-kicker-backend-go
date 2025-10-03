package user

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeCreate(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateUserRequest(request.(*entities.CreateUserRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, err
		}

		id, err := s.Create(ctx, *request.(*entities.CreateUserRequest))
		if err != nil {
			return nil, err
		}

		response := &struct {
			Id int64 `json:"id"`
		}{}

		response.Id = *id

		return response, nil
	}
}

func makeLogin(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeLogin").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.LoginUserRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		uID, uRole, uTeamID, token, cityId, err := s.Login(ctx, req)
		if err != nil {
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &entities.LoginUserResponse{
			ID:     *uID,
			Token:  *token,
			Role:   *uRole,
			TeamID: uTeamID,
			CityID: cityId,
		}

		return response, nil
	}
}

func makeChangePassword(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastRequest[*entities.ChangePasswordRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateChangePasswordRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, err
		}

		userOld := &entities.User{
			Email:    req.Email,
			Password: []byte(req.Password),
		}

		userNew := &entities.User{
			Email:    req.Email,
			Password: []byte(req.NewPassword),
		}

		err = s.ChangePassword(ctx, *userOld, *userNew)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func makeCheckAuth(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		logger := s.GetLogger().With().Str("Source", "makeCheckAuth").Logger()

		req, err := helpers.CastRequest[*entities.CheckAuthRequest](request)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, err
		}

		isAuth, role, teamID, cityId, err := s.CheckAuth(ctx, req.UserID, req.Token)
		if err != nil {
			return nil, err
		}

		response := &entities.CheckAuthResponse{
			IsAuthenticated: *isAuth,
			Role:            *role,
			TeamID:          *teamID,
			CityID:          cityId,
		}

		return response, nil
	}
}

func makeCreateTournamentMaster(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreateTournamentMaster").Logger()

		req, err := helpers.CastRequest[*entities.CreateTournamentMasterRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		id, err := s.CreateTournamentMaster(ctx, *request.(*entities.CreateTournamentMasterRequest))
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeCreateTournamentMaster")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return &struct {
			Id int64 `json:"id"`
		}{
			Id: id,
		}, nil
	}
}

func makeUpdateTournamentMaster(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdateTournamentMaster").Logger()

		req, err := helpers.CastRequest[*entities.UpdateTournamentMasterRequest](request)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeUpdateTournamentMaster")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.UpdateTournamentMaster(ctx, *req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeUpdateTournamentMaster")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct{}{}, nil
	}
}

func makeGetTournamentMasterListByCityId(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentMasterListByCityId").Logger()

		id, err := helpers.CastRequest[int64](request)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		if id <= 0 {
			err = errors.New(pkgerr.WrongParameterError)
			serviceLogger.Error().Err(err).Msg("failed user.makeGetTournamentMasterListByCityId")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		tMasters, err := s.GetTournamentMasterListByCityId(ctx, id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeGetTournamentMasterListByCityId")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.GetTournamentMastersResponse{
			Masters: tMasters,
		}, nil
	}
}

func makeGetTournamentMasterUserById(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentMasterUserById").Logger()

		id, err := helpers.CastRequest[int64](request)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		if id <= 0 {
			err = errors.New(pkgerr.WrongParameterError)
			serviceLogger.Error().Err(err).Msg("failed user.makeGetTournamentMasterUserById")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		tMaster, err := s.GetTournamentMasterByUserId(ctx, id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeGetTournamentMasterUserById")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.GetTournamentMastersResponse{
			Masters: []entities.TournamentMaster{tMaster},
		}, nil
	}
}

func makeGetTournamentMasterList(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentMasterList").Logger()

		tMasters, err := s.GetTournamentMasterList(ctx)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeGetTournamentMasterList")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return entities.GetTournamentMastersResponse{
			Masters: tMasters,
		}, nil
	}
}

func makeDeleteTournamentMaster(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeDeleteTournamentMaster").Logger()

		id, err := helpers.CastRequest[int64](request)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(pkgerr.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		if id <= 0 {
			err = errors.New(pkgerr.WrongParameterError)
			serviceLogger.Error().Err(err).Msg("failed user.makeDeleteTournamentMaster")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.DeleteTournamentMaster(ctx, id)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed user.makeDeleteTournamentMaster")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return struct{}{}, nil
	}
}
