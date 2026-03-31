package team

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/team"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"
)

func makeGetTeam(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeam").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateGetTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response, err := s.GetTeam(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTeam")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return response, nil
	}
}

func makeGetTeams(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeams").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamsRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResponse, err := s.GetTeams(ctx, req.CityId, req.OnlyFree)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTeams")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return teamsResponse, nil
	}
}

func makeGetTeamsByCity(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamsByCity").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamsByCityRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResponse, err := s.GetTeamsByCity(ctx, req.OnlyFree, req.CityID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTeamsByCity")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return teamsResponse, nil
	}
}

func makeGetTeamsByLeague(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamsByLeague").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamsByLeagueRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResponse, err := s.GetTeamsByLeague(ctx, req.LeagueID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTeamsByLeague")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return teamsResponse, nil
	}
}

func makeGetTeamVsTeamTable(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamVsTeamTable").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamVsTeamTableRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResp, err := s.GetTeamVsTeamTable(ctx, req.CityID, req.SeasonID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTeamVsTeamTable")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		// TLeaguesTables
		return teamsResp, nil
	}
}

func makeGetTournamentTeamVsTeamTable(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentTeamVsTeamTable").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamVsTeamTableRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResp, err := s.GetTournamentTeamVsTeamTable(ctx, req.CityID, req.SeasonID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTournamentTeamVsTeamTable")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return teamsResp, nil
	}
}

func makeGetTournamentsTeamVsTeamTable(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTournamentsTeamVsTeamTable").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.GetTeamVsTeamTableRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		teamsResp, err := s.GetTournamentsTeamVsTeamTable(ctx, req.CityID, req.SeasonID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.GetTournamentsTeamVsTeamTable")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		return teamsResp, nil
	}
}

func makeCreate(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.CreateTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		id, err := s.Create(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.Create")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := struct {
			Id int64 `json:"id"`
		}{}

		response.Id = id

		return response, nil
	}
}

func makeUpdate(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.UpdateTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.Update(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.makeUpdate")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

func makeDelete(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makemakeDelete").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.DeleteTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateDeleteTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		res, err := s.Delete(ctx, req.ID)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.Delete")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

func makeAddPlayerIntoTeam(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeAddPlayerIntoTeam").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.MovingPlayerTeam](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.AddPlayerIntoTeam(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.AddPlayerIntoTeam")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}

func makeRemovePlayerFromTeam(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeRemovePlayerFromTeam").Str("request_id", reqID).Logger()

		req, err := helpers.CastRequest[*entities.MovingPlayerTeam](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg(errors.FailedCastRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = s.GetValidator().Struct(req)
		if err != nil {
			serviceLogger.Error().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		res, err := s.RemovePlayerFromTeam(ctx, req)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("failed to team.RemovePlayerFromTeam")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		response := &struct {
			Success bool `json:"success"`
		}{}
		response.Success = res

		return response, nil
	}
}
