package game

import (
	"context"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
)

func (s *Service) Create(ctx context.Context, request entities.CreateGameRequest) (entities.CreateGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Create").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rwdbOperations.CreatePlayedGame(logger, timeout, request)
	if err != nil {
		return entities.CreateGameResponse{}, err
	}
	return resp, nil
}

func (s *Service) Delete(ctx context.Context, gameID int) error {
	logger := s.logger.With().Interface("service", "game.Delete").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.DeleteGame(logger, timeout, gameID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Get(ctx context.Context, gameID int) (entities.GetGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Get").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rdbOperations.GetGame(logger, timeout, gameID)
	if err != nil {
		return entities.GetGameResponse{}, err
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, request entities.UpdateGameRequest) error {
	logger := s.logger.With().Interface("service", "game.Update").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.UpdateGame(logger, timeout, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Find(ctx context.Context, request entities.FindGameRequest) (entities.FindGameResponse, error) {
	logger := s.logger.With().Interface("service", "game.Find").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.FindGames(logger, timeout, request)
	if err != nil {
		return entities.FindGameResponse{}, err
	}

	return entities.FindGameResponse{
		Games: games,
	}, nil
}

func (s *Service) UpdateFutureGame(ctx context.Context, request entities.UpdateFutureGameRequest) error {
	logger := s.logger.With().Interface("service", "game.UpdateFutureGame").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RWDB.MaxIdleConnectionTimeout)
	defer cancel()

	err := s.rwdbOperations.UpdateFutureGame(logger, timeout, request)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetGamesYears(ctx context.Context) (entity.GetGamesYearsResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetGamesYears").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	resp, err := s.rdbOperations.GetGamesYears(logger, timeout)
	if err != nil {
		return entity.GetGamesYearsResponse{}, err
	}
	return resp, nil
}

func (s *Service) GetComingGames(ctx context.Context) (entities.GetComingGamesResponse, error) {
	logger := s.logger.With().Interface("service", "game.GetComingGames").Logger()
	timeout, cancel := context.WithTimeout(ctx, s.config.RDB.MaxIdleConnectionTimeout)
	defer cancel()

	games, err := s.rdbOperations.GetComingGames(logger, timeout)
	if err != nil {
		return entities.GetComingGamesResponse{}, err
	}

	return entities.GetComingGamesResponse{ComingGames: games}, nil
}
