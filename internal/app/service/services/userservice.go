package services

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/service"
	"back-end/internal/app/store"
	"context"
)

type UserService struct {
	service *Service
}

func (s *UserService) Create(user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	return s.service.store.User().Create(user)
}

func (s *UserService) GetAllUsers(limit, offset int) ([]models.User, error) {
	return s.service.store.User().GetAll(limit, offset)
}

// Find returns *models.User by userID or nil if user not found.
//	If user was not found, method returns service.ErrUserNotFound.
//	If an error occurs during the execution of the method, the method returns error.
func (s *UserService) Find(userID int) (*models.User, error) {
	u, err := s.service.store.User().Find(userID)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrUserNotFound
	}

	return u, nil
}

func (s *UserService) IsUserExist(ctx context.Context, userID int) (bool, error) {
	return s.service.store.User().IsUserExist(userID)
}

func (s *UserService) FindByEmail(email string) (*models.User, error) {
	if "" != email {
		u, err := s.service.store.User().FindByEmail(email)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrUserNotFound
		}

		return u, nil
	}
	return nil, service.ErrInvalidUserEmail
}

func (s *UserService) FindByLogin(login string) (*models.User, error) {
	if "" != login {
		u, err := s.service.store.User().FindByLogin(login)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrUserNotFound
		}

		return u, nil
	}
	return nil, service.ErrInvalidUserLogin
}
