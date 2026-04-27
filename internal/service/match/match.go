package match

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
)

func (s *Service) Create(ctx context.Context, match entities.Match) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	if err := s.validateTeams(ctx, logger, match); err != nil {
		return nil, err
	}

	id, err := s.rwdbOperations.CreateMatch(logger, ctx, match)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (s *Service) Update(ctx context.Context, match entities.Match) error {
	logger := s.logger.With().Interface("service", "Update").Logger()

	oldMatch, err := s.rdbOperations.GetMatchById(logger, ctx, match.ID)
	if err != nil {
		return err
	}

	validateMatch := match
	if match.GameID == nil { // если в запросе нет match.GameID, то проверим в бд существующий oldMatch.GameID
		validateMatch.GameID = oldMatch.GameID
	}

	if err := s.validateTeams(ctx, logger, validateMatch); err != nil {
		return err
	}

	team1Id := oldMatch.Team1ID
	team2Id := oldMatch.Team2ID

	if match.Team1ID != nil {
		team1Id = match.Team1ID
	}
	if match.Team2ID != nil {
		team2Id = match.Team2ID
	}

	if (team1Id != nil && team2Id != nil) &&
		(*team1Id == *team2Id) {
		msg := fmt.Sprintf("в матче получаются одинаковые команды (%d)", *team1Id)
		logger.Error().Msg(msg)
		return error_templates.BadRequestError(errors.New(msg))
	}

	err = s.rwdbOperations.UpdateMatch(logger, ctx, match)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) (bool, error) {
	logger := s.logger.With().Interface("service", "Delete").Logger()

	res, err := s.rwdbOperations.DeleteMatch(logger, ctx, id)
	if err != nil {
		return res, err
	}

	return res, nil
}

// validateTeams проверяет соответствие команд игры при наличии GameID
func (s *Service) validateTeams(ctx context.Context, logger zerolog.Logger, match entities.Match) error {
	if match.GameID == nil || (match.Team1ID == nil && match.Team2ID == nil) {
		return nil
	}

	if (match.Team1ID != nil && match.Team2ID != nil) &&
		(*match.Team1ID == *match.Team2ID) {
		msg := "указаны одинаковые команды"
		logger.Error().Msg(msg)
		return error_templates.BadRequestError(errors.New(msg))
	}

	game, err := s.rdbOperations.GetGameById(logger, ctx, int(*match.GameID), nil)
	if err != nil {
		return err
	}

	// Проверяем Team1ID если он передан
	if match.Team1ID != nil && (game.Team1ID != *match.Team1ID && game.Team2ID != *match.Team1ID) {
		msg := fmt.Sprintf("team1Id (%d) матча не соответствует командам игры (%d,%d)", *match.Team1ID, game.Team1ID, game.Team2ID)
		logger.Error().Msg(msg)
		return error_templates.BadRequestError(errors.New(msg))
	}

	// Проверяем Team2ID если он передан
	if match.Team2ID != nil && (game.Team1ID != *match.Team2ID && game.Team2ID != *match.Team2ID) {
		msg := fmt.Sprintf("team2Id (%d) матча не соответствует командам игры (%d,%d)", *match.Team2ID, game.Team1ID, game.Team2ID)
		logger.Error().Msg(msg)
		return error_templates.BadRequestError(errors.New(msg))
	}

	return nil
}
