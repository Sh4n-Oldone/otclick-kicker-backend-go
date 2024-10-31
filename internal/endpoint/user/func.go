package user

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	// "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/user"
)

func makeCreate(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		err := helpers.ValidateCreateUserRequest(request.(*entity.CreateUserRequest))
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		id, err := s.Create(ctx, *request.(*entity.CreateUserRequest))
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
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeLogin").Logger()

		req, err := helpers.CastRequest[*entity.LoginUserRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		user := helpers.ConvertLoginUserRequestToUser(req)

		uID, uRole, uTeamID, token, err := s.Login(ctx, *user)
		if err != nil {
			return nil, err
		}

		response := &entity.LoginUserResponse {
			Message: "Access Granted",
			ID: *uID,
			Token: *token,
			Role: *uRole,
			TeamID: *uTeamID,
		}

		return response, nil
	}
}

func makeChangePassword(s user.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		// reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastRequest[*entity.ChangePasswordRequest](request)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		err = helpers.ValidateChangePasswordRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		userOld := &entity.User{
			Email:    req.Email,
			Password: []byte(req.Password),
		}

		userNew := &entity.User{
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

		req, err := helpers.CastRequest[*entity.CheckAuthRequest](request)
		if err != nil {
			logger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedCastRequest)
			return nil, err
		}

		isAuth, role, teamID, err := s.CheckAuth(ctx, req.UserID, req.Token)
		if err != nil {
			return nil, err
		}

		response := &entity.CheckAuthResponse{
			IsAuthenticated: *isAuth,
			Role: *role,
			TeamID: *teamID,
		}

		return response, nil
	}
}

