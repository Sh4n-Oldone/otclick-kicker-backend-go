package user

import (
	"context"

	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/internal/entity"
	"node71.otclick.ru/sideprojects/kicker/kicker-backend-go/pkg/crypt"
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

func (s *Service) CheckAuth(ctx context.Context) error {
	return nil
}

func (s *Service) GetUser(ctx context.Context, userID int64) (*entity.User, error) {
	logger := s.logger.With().Interface("service", "GetUser").Logger()

	user, err := s.rdbOperations.GetUser(logger, ctx, &userID, nil)
	if err != nil {
		return nil, err
	}

	return user, nil
}
