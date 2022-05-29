package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
)

type UserService struct {
	service *Service
}

func (s *UserService) Create(ctx context.Context, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return s.service.store.User().Create(ctx, user)
}

func (s *UserService) GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	return s.service.store.User().GetAll(ctx, limit, offset)
}

// Find returns *models.User by userID or nil if user not found.
//	If user was not found, method returns service.ErrUserNotFound.
//	If an error occurs during the execution of the method, the method returns error.
func (s *UserService) Find(ctx context.Context, userID int) (*models.User, error) {
	u, err := s.service.store.User().Find(ctx, userID)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrUserNotFound
	}

	return u, nil
}

func (s *UserService) IsUserExist(ctx context.Context, userID int) (bool, error) {
	return s.service.store.User().IsUserExist(ctx, userID)
}

func (s *UserService) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if "" != email {
		u, err := s.service.store.User().FindByEmail(ctx, email)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrUserNotFound
		}

		return u, nil
	}
	return nil, service.ErrInvalidUserEmail
}

func (s *UserService) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	if "" != login {
		u, err := s.service.store.User().FindByLogin(ctx, login)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrUserNotFound
		}

		return u, nil
	}
	return nil, service.ErrInvalidUserLogin
}
