package team

import (
	"context"

	"github.com/go-kit/kit/endpoint"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/transport/http/middleware"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/helpers"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/team"
)

// GetTeam {id}
// GetTeams
// GetTeamsByCity {city_id}
// GetTeamsByLeague {league_id}
// GetTeamVsTeamTable {city_id}

// Create
// Update
// Delete {id}

// AddPlayerIntoTeam
// RemovePlayerFromTeam
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// func makeGetTeam(s team.IService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		reqID, ctx := middleware.GetRequestID(ctx)
// 		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeam").Logger()

// 		req, err := helpers.CastRequest[*entity.GetTeamRequest](request)
// 		if err != nil {
// 			serviceLogger.Error().Err(err).Msg("Failed to cast request")
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		err = helpers.ValidateGetTeamRequest(req)
// 		if err != nil {
// 			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		response, err := s.GetTeam(ctx, req.ID)
// 		if err != nil {
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}
// 		// TeamFilledWithFullPlayers
// 		//GetTeamResponse
// 		return response, nil
// 	}
// }

// func makeGetTeams(s team.IService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		reqID, ctx := middleware.GetRequestID(ctx)
// 		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeams").Logger()

// 		err := helpers.ValidateGetTeamsRequest(request.(*entity.GetTeamsRequest))
// 		if err != nil {
// 			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		teamReq := helpers.ConvertGetTeamsRequestToTeam(request.(*entity.GetTeamsRequest))

// 		teamsResp, err := s.GetTeams(ctx, *teamReq)
// 		if err != nil {
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}
// 		// id
// 		// teamName
// 		// teamShortName
// 		return teamsResp, nil
// 	}
// }

// 3
// func makeGetTeamsByCity(s team.IService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		reqID, ctx := middleware.GetRequestID(ctx)
// 		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamsByCity").Logger()

// 		err := helpers.ValidateGetTeamsByCityRequest(request.(*entity.GetTeamsByCityRequest))
// 		if err != nil {
// 			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		teamReq := helpers.ConvertGetTeamsByCityRequestToTeam(request.(*entity.GetTeamsByCityRequest))

// 		teamsResp, err := s.GetTeamsByCity(ctx, *teamReq)
// 		if err != nil {
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}
// 		// id
// 		// teamName
// 		// teamShortName
// 		return teamsResp, nil
// 	}
// }

// 4
// func makeGetTeamsByLeague(s team.IService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		reqID, ctx := middleware.GetRequestID(ctx)
// 		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamsByLeague").Logger()

// 		err := helpers.ValidateGetTeamsByLeagueRequest(request.(*entity.GetTeamsByLeagueRequest))
// 		if err != nil {
// 			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		teamReq := helpers.ConvertGetTeamsByLeagueRequestToTeam(request.(*entity.GetTeamsByLeagueRequest))

// 		teamsResp, err := s.GetTeamsByLeague(ctx, *teamReq)
// 		if err != nil {
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}
// 		// type Team = {
// 		// 	name: string;
// 		// 	shortName: string;
// 		// 	avatar?: string | undefined;
// 		// 	cityId: number;
// 		// 	players: Player[];
// 		// }
// 		return teamsResp, nil
// 	}
// }

// 5
// func makeGetTeamVsTeamTable(s team.IService) endpoint.Endpoint {
// 	return func(ctx context.Context, request interface{}) (interface{}, error) {
// 		reqID, ctx := middleware.GetRequestID(ctx)
// 		serviceLogger := s.GetLogger().With().Str("Source", "makeGetTeamVsTeamTable").Logger()

// 		err := helpers.ValidateGetTeamVsTeamTableRequest(request.(*entity.GetTeamVsTeamTableRequest))
// 		if err != nil {
// 			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		teamReq := helpers.ConvertGetTeamVsTeamTableRequestToTeam(request.(*entity.GetTeamVsTeamTableRequest))

// 		teamsResp, err := s.GetTeamVsTeamTable(ctx, *teamReq)
// 		if err != nil {
// 			return nil, error_templates.WrapErrorEndpoint(err, reqID)
// 		}

// 		// TLeaguesTables
// 		return teamsResp, nil
// 	}
// }

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func makeCreate(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeCreate").Logger()

		req, err := helpers.CastRequest[*entity.CreateTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateCreateTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		id, err := s.Create(ctx, *req)
		if err != nil {
			return nil, err
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
		serviceLogger := s.GetLogger().With().Str("Source", "makeUpdate").Logger()

		req, err := helpers.CastRequest[*entity.UpdateTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateUpdateTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		response, err := s.Update(ctx, *req)
		if err != nil {
			return response, err
		}

		return response, nil
	}
}

func makeDelete(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makemakeDelete").Logger()

		req, err := helpers.CastRequest[*entity.DeleteTeamRequest](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return nil, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidateDeleteTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return nil, err
		}

		response, err := s.Delete(ctx, req.ID)
		if err != nil {
			return response, err
		}

		return response, nil
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func makeAddPlayerIntoTeam(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeAddPlayerIntoTeam").Logger()

		req, err := helpers.CastRequest[*entity.PlayerTeam](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidatePlayerTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, err
		}

		response, err := s.AddPlayerIntoTeam(ctx, req.PlayerID, req.TeamID)
		if err != nil {
			return response, err
		}

		return response, nil
	}
}

func makeRemovePlayerFromTeam(s team.IService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		reqID, ctx := middleware.GetRequestID(ctx)
		serviceLogger := s.GetLogger().With().Str("Source", "makeRemovePlayerFromTeam").Logger()

		req, err := helpers.CastRequest[*entity.PlayerTeam](request)
		if err != nil {
			serviceLogger.Error().Err(err).Msg("Failed to cast request")
			return false, error_templates.WrapErrorEndpoint(err, reqID)
		}

		err = helpers.ValidatePlayerTeamRequest(req)
		if err != nil {
			serviceLogger.Error().Stack().Err(error_templates.ErrorDetailFromError(err)).Msg(errors.FailedValidateRequest)
			return false, err
		}

		response, err := s.RemovePlayerFromTeam(ctx, req.PlayerID, req.TeamID)
		if err != nil {
			return response, err
		}

		return response, nil
	}
}
