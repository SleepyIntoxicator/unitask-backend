package teststore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
)

type UserRepository struct {
	store *Store
	users map[int]*models.User //Map[user.email]*User
}

func (r *UserRepository) Create(u *models.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	u.ID = len(r.users) + 1
	r.users[u.ID] = u

	return nil
}

func (r *UserRepository) Find(id int) (*models.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, store.ErrRecordNotFound
	}

	return u, nil
}

func (r *UserRepository) FindByLogin(login string) (*models.User, error) {
	for _, u := range r.users {
		if u.Login == login {
			return u, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}

	return nil, store.ErrRecordNotFound
}

func (r *UserRepository) CreateTester() (*models.User, error) {
	u := &models.User{
		Login:    "tester",
		FullName: "Tester tester",
		Email:    "Tester@example.org",
		Password: "Testerpassword",
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	u.ID = len(r.users) + 1
	r.users[u.ID] = u

	return u, nil
}

func (r *UserRepository) GetUserToken(userID int) (*models.UserToken, error) {
	panic("implement me")
}

func (UserRepository) GetAll(limit, offset int) ([]models.User, error) {
	panic("implement me")
}

func (UserRepository) IsUserExist(userID int) (bool, error) {
	panic("implement me")
}

func (UserRepository) GetUserRoles(userID int) ([]models.Role, error) {
	panic("implement me")
}
