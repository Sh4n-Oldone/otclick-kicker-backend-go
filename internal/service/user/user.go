package user

import (
	"context"
	stderr "errors"
	"net/http"
	"time"

	jwtV5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"

	cnst "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/constant"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/crypt"
	errTmpl "node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/error_templates"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/errors"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/jwt"
)

func (s *Service) Create(ctx context.Context, request entity.CreateUserRequest) (*int64, error) {
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

	user := &entity.User{
		Email:    request.Email,
		Password: passHash,
		Role:     role,
	}

	id, err := s.rwdbOperations.CreateUser(logger, ctx, *user)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (s *Service) Login(ctx context.Context, user entity.User) (*string, error) {
	logger := s.logger.With().Interface("service", "Login").Logger()

	_user, err := s.rdbOperations.GetUser(logger, ctx, nil, &user.Email)
	if err != nil {
		return nil, err
	}

	err = crypt.ComparePasswordAndHash([]byte(user.Password), []byte(_user.Password), []byte(s.config.Secret.Salt))
	if err != nil {
		return nil, err
	}

	token, err := jwt.NewToken(*_user, s.config.Token.AccessTTL, s.config.Secret.Key)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (s *Service) ChangePassword(ctx context.Context, userOld, userNew entity.User) error {
	logger := s.logger.With().Interface("service", "Login").Logger()

	_user, err := s.rdbOperations.GetUser(logger, ctx, nil, &userOld.Email)
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

func (s *Service) CheckAuth(ctx context.Context, userID int64, token string) (*bool, *string, *int64, error) {
	logger := s.logger.With().Interface("service", "CheckAuth").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, &userID, nil)
	if err != nil {
		return nil, nil, nil, errTmpl.New(errors.FailedGetUserData, err, codes.InvalidArgument, http.StatusBadRequest)
	}

	_token, err := jwtV5.Parse(token, func(token *jwtV5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtV5.SigningMethodHMAC); !ok {
			return nil, jwtV5.ErrSignatureInvalid
		}
		return []byte(s.config.Secret.Key), nil
	})
	if err != nil {
		return nil, nil, nil, errTmpl.New(err.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
	}

	if claims, ok := _token.Claims.(jwtV5.MapClaims); ok && _token.Valid {
		stdErr := stderr.New(errors.WrongParameterError)
		err = errTmpl.New(stdErr.Error(), err, codes.InvalidArgument, http.StatusBadRequest)
		// error if Access-Token is expired
		expired, ok2 := claims[cnst.JwtClaimsAttrTokenExpire].(float64)
		if !ok2 {
			return nil, nil, nil, err
		}
		if expired < float64(time.Now().Unix()) {
			return nil, nil, nil, err
		}

		// get user ID
		userID, ok2 := claims[cnst.JwtClaimsAttrUserID].(float64)
		if !ok2 {
			return nil, nil, nil, err
		}
		
		// get user Email
		ok2 = false
		userEmail, ok2 := claims[cnst.JwtClaimsAttrUserEmail]
		if !ok2 {
			return nil, nil, nil, err
		}

		// get user Role name
		ok2 = false
		role, ok2 := claims[cnst.JwtClaimsAttrRoleName]
		if !ok2 {
			return nil, nil, nil, err
		}

		// get user team ID
		ok2 = false
		teamID, ok2 := claims[cnst.JwtClaimsAttrTeamID].(float64)
		if !ok2 {
			return nil, nil, nil, err
		}

		if user.ID != int64(userID) || user.Email != userEmail || user.Role.Name != role || user.Team.ID != int64(teamID) {
			return nil, nil, nil, err
		}
	}	

	auth := true

	return &auth, &user.Role.Name, &user.Team.ID, nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (*entity.User, error) {
	logger := s.logger.With().Interface("service", "GetUser").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, &userID, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}
