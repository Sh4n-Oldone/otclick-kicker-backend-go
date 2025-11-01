package place

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
)

func (s *Service) GetList(ctx context.Context, barID, tableID, cityID *int64, withDeleted bool) ([]entities.Place, error) {
	logger := s.logger.With().Str("service", "GetPlaceList").Logger()

	entities, err := s.rdbOperations.GetPlaceList(logger, ctx, barID, tableID, cityID, withDeleted)
	if err != nil {
		return nil, err
	}

	return entities, nil
}

func (s *Service) Get(ctx context.Context, id int64) (*entities.Place, error) {
	logger := s.logger.With().Str("service", "GetPlace").Logger()

	entity, err := s.rdbOperations.GetPlaceByID(logger, ctx, id, &s.config.RDB)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *Service) Create(ctx context.Context, request entities.CreatePlaceRequest) (*int64, error) {
	logger := s.logger.With().Str("service", "Create").Logger()

	if request.Creator.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Creator.ID)
		if err != nil {
			return nil, err
		}

		bar, err := s.rdbOperations.GetBarByID(logger, ctx, request.BarID, &s.config.RDB)
		if err != nil {
			return nil, err
		}

		if master.City.ID != bar.City.ID {
			err = errors.New(pkgerr.ErrBarNotInMasterCity)
			return nil, error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	id, err := s.rwdbOperations.CreatePlace(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, request entities.UpdatePlaceRequest) error {
	logger := s.logger.With().Str("service", "Update").Logger()

	if request.Executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, request.Executor.ID)
		if err != nil {
			return err
		}

		bar, err := s.rdbOperations.GetBarByID(logger, ctx, request.BarID, &s.config.RDB)
		if err != nil {
			return err
		}

		if master.City.ID != bar.City.ID {
			err = errors.New(pkgerr.ErrBarNotInMasterCity)
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	err := s.rwdbOperations.UpdatePlace(logger, ctx, request, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64, executor entities.User) error {
	logger := s.logger.With().Str("service", "Delete").Logger()

	if executor.Role.Name == constant.TournamentMaster {
		master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, executor.ID)
		if err != nil {
			return err
		}

		place, err := s.rdbOperations.GetPlaceByID(logger, ctx, id, &s.config.RDB)
		if err != nil {
			return err
		}

		bar, err := s.rdbOperations.GetBarByID(logger, ctx, place.Bar.ID, &s.config.RDB)
		if err != nil {
			return err
		}

		if master.City.ID != bar.City.ID {
			err = errors.New(pkgerr.ErrBarNotInMasterCity)
			return error_templates.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		}
	}

	err := s.rwdbOperations.DeletePlace(logger, ctx, id, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}
