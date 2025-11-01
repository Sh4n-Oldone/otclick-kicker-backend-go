package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwtV5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/service/entities"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/crypt"
	errTmpl "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	pkgerr "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/jwt"
)

func (s *Service) Create(ctx context.Context, request entities.CreateUserRequest) (*int64, error) {
	logger := s.logger.With().Interface("service", "Create").Logger()

	// check Role exist
	role, err := s.rdbOperations.GetRole(logger, ctx, nil, &request.RoleName)
	if err != nil {
		return nil, err
	}

	passHash, err := crypt.EncryptPassword([]byte(request.Password), []byte(s.config.Secret.Salt))
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Email:    request.Email,
		Password: passHash,
		Role:     role,
	}

	if request.TeamID != 0 {
		user.Team = &entities.Team{ID: request.TeamID}
	}

	id, err := s.rwdbOperations.CreateUser(logger, ctx, *user)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, req *entities.LoginUserRequest) (*int64, *string, *int64, *string, *int64, error) {
	logger := s.logger.With().Str("service", "Login").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, nil, &req.Email, &s.config.RDB)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	var cityId *int64

	// город можем получить пока толко у мастера, остальные роли в таблицу users_cities_links не заносятся
	if user.Role != nil && user.Role.Name == cnst.TournamentMaster {
		city, err := s.rdbOperations.GetUserCity(logger, ctx, user.ID)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}

		cityId = &city.ID
	}

	err = crypt.ComparePasswordAndHash([]byte(req.Password), user.Password, []byte(s.config.Secret.Salt))
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	token, err := jwt.NewToken(*user, s.config.Token.AccessTTL, s.config.Secret.Key)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return &user.ID, &user.Role.Name, &user.Team.ID, &token, cityId, nil
}

func (s *Service) ChangePassword(ctx context.Context, userOld, userNew entities.User) error {
	logger := s.logger.With().Interface("service", "Login").Logger()

	_user, err := s.rdbOperations.GetUser(logger, ctx, nil, &userOld.Email, &s.config.RDB)
	if err != nil {
		return err
	}

	err = crypt.ComparePasswordAndHash([]byte(userOld.Password), []byte(_user.Password), []byte(s.config.Secret.Salt))
	if err != nil {
		return err
	}

	passHash, err := crypt.EncryptPassword([]byte(userNew.Password), []byte(s.config.Secret.Salt))
	if err != nil {
		return err
	}

	_user.Password = passHash

	err = s.rwdbOperations.UpdateUser(logger, ctx, *_user)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CheckAuth(ctx context.Context, userID int64, token string) (*bool, *string, *int64, *int64, error) {
	logger := s.logger.With().Interface("service", "CheckAuth").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, &userID, nil, &s.config.RDB)
	if err != nil {
		return nil, nil, nil, nil, errTmpl.New(pkgerr.FailedGetUserData, err, codes.InvalidArgument, http.StatusBadRequest)
	}

	var cityId *int64

	if user.Role != nil && user.Role.Name == cnst.TournamentMaster {
		city, err := s.rdbOperations.GetUserCity(logger, ctx, user.ID)
		if err != nil {
			return nil, nil, nil, nil, errTmpl.New(pkgerr.FailedGetUserData, err, codes.InvalidArgument, http.StatusBadRequest)
		}

		cityId = &city.ID
	}

	_token, err := jwtV5.Parse(token, func(token *jwtV5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtV5.SigningMethodHMAC); !ok {
			return nil, jwtV5.ErrSignatureInvalid
		}
		return []byte(s.config.Secret.Key), nil
	})
	if err != nil {
		return nil, nil, nil, nil, errTmpl.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if claims, ok := _token.Claims.(jwtV5.MapClaims); ok && _token.Valid {
		stdErr := errors.New(pkgerr.WrongParameterError)
		err = errTmpl.New(stdErr.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		// error if Access-Token is expired
		expired, ok2 := claims[cnst.JwtClaimsAttrTokenExpire].(float64)
		if !ok2 {
			return nil, nil, nil, nil, err
		}
		if expired < float64(time.Now().Unix()) {
			return nil, nil, nil, nil, err
		}

		// get user ID
		userID, ok2 := claims[cnst.JwtClaimsAttrUserID].(float64)
		if !ok2 {
			return nil, nil, nil, nil, err
		}

		// get user Email
		ok2 = false
		userEmail, ok2 := claims[cnst.JwtClaimsAttrUserEmail]
		if !ok2 {
			return nil, nil, nil, nil, err
		}

		// get user Role name
		ok2 = false
		role, ok2 := claims[cnst.JwtClaimsAttrRoleName]
		if !ok2 {
			return nil, nil, nil, nil, err
		}

		// get user team ID
		ok2 = false
		teamID, ok2 := claims[cnst.JwtClaimsAttrTeamID].(float64)
		if !ok2 {
			return nil, nil, nil, nil, err
		}

		if user.ID != int64(userID) || user.Email != userEmail || user.Role.Name != role || user.Team.ID != int64(teamID) {
			return nil, nil, nil, nil, err
		}
	}

	auth := true

	return &auth, &user.Role.Name, &user.Team.ID, cityId, nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (*entities.User, error) {
	logger := s.logger.With().Interface("service", "GetUser").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, &userID, nil, &s.config.RDB)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) CreateTournamentMaster(ctx context.Context, request entities.CreateTournamentMasterRequest) (int64, error) {
	logger := s.logger.With().Str("service", "CreateTournamentMaster").Logger()

	// не даем создать мастера с удалённым городом
	city, err := s.rdbOperations.GetCityById(logger, ctx, request.CityID)
	if err != nil {
		return 0, err
	}
	if city.DeletedAt != nil {
		err = fmt.Errorf("город мастера \"%s\" удалён %v", city.Ru, city.DeletedAt.Format(time.DateOnly))
	}

	role, err := s.rdbOperations.GetRole(logger, ctx, nil, &request.RoleName)
	if err != nil {
		return 0, err
	}

	passHash, err := crypt.EncryptPassword([]byte(request.Password), []byte(s.config.Secret.Salt))
	if err != nil {
		return 0, errTmpl.New(err.Error(), err, codes.Internal, http.StatusInternalServerError)
	}

	tMaster := entities.TournamentMaster{
		User: entities.User{
			Email:    request.Email,
			Password: passHash,
			Role:     role,
		},
		City: entities.City{
			ID: request.CityID,
		},
	}

	id, err := s.rwdbOperations.CreateTournamentMaster(logger, ctx, tMaster, &s.config.RWDB)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) UpdateTournamentMaster(ctx context.Context, request entities.UpdateTournamentMasterRequest) error {
	logger := s.logger.With().Str("service", "UpdateTournamentMaster").Logger()

	if request.CityID != nil {
		err := s.rwdbOperations.UpdateTournamentMaster(logger, ctx, request, &s.config.RWDB)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetTournamentMasterListByCityId(ctx context.Context, cityID int64) ([]entities.TournamentMaster, error) {
	logger := s.logger.With().Str("service", "GetTournamentMasterListByCityId").Logger()

	tMasters, err := s.rdbOperations.GetTournamentMasterListByCityId(logger, ctx, cityID, &s.config.RDB)
	if err != nil {
		return nil, err
	}

	return tMasters, nil
}

func (s *Service) GetTournamentMasterByUserId(ctx context.Context, userID int64) (entities.TournamentMaster, error) {
	logger := s.logger.With().Str("service", "GetTournamentMasterByUserId").Logger()

	tMaster, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, userID)
	if err != nil {
		return entities.TournamentMaster{}, err
	}

	return tMaster, nil
}

func (s *Service) GetTournamentMasterList(ctx context.Context) ([]entities.TournamentMaster, error) {
	logger := s.logger.With().Str("service", "GetTournamentMasterList").Logger()

	tMasters, err := s.rdbOperations.GetTournamentMasterList(logger, ctx, &s.config.RDB)
	if err != nil {
		return nil, err
	}

	return tMasters, nil
}

func (s *Service) DeleteTournamentMaster(ctx context.Context, userID int64) error {
	logger := s.logger.With().Str("service", "DeleteTournamentMaster").Logger()

	master, err := s.rdbOperations.GetTournamentMasterByUserId(logger, ctx, userID)
	if err != nil {
		return err
	}

	if master.User.Role == nil || master.User.Role.Name != cnst.TournamentMaster {
		err = errors.New("only for role " + cnst.TournamentMaster)
		logger.Error().Err(err).Msg("failed user.DeleteTournamentMaster")
		return errTmpl.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	err = s.rwdbOperations.DeleteTournamentMaster(logger, ctx, userID, &s.config.RWDB)
	if err != nil {
		return err
	}

	return nil
}
