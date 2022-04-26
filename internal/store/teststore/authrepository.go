package teststore

import (
	"back-end/internal/app/api/v1/models"
	"github.com/google/uuid"
)

type AuthRepository struct {
	store *Store
	users map[int]*models.User	//Map[user.email]*User
}

func (r *AuthRepository) RegisterApp(app *models.RegisteredApp) error {
	panic("implement me")
}

func (r *AuthRepository) DeleteApp(appUUID uuid.UUID) error {
	panic("implement me")
}

func (r *AuthRepository) GetApp(appUUID uuid.UUID) (*models.RegisteredApp, error) {
	panic("implement me")
}

func (r *AuthRepository) GetAppByName(name string) (*models.RegisteredApp, error) {
	panic("implement me")
}

func (r *AuthRepository) AddAppToken(t *models.AppToken) error {
	panic("implement me")
}

func (r *AuthRepository) RemoveAppTokens(appUUID uuid.UUID) error {
	panic("implement me")
}

func (r *AuthRepository) GetAppTokenInfo(token string) (*models.AppToken, error) {
	panic("implement me")
}

func (r *AuthRepository) GetAppTokenByAppUUID(appUUID uuid.UUID) (*models.AppToken, error) {
	panic("implement me")
}

func (r *AuthRepository) AddUserToken(t *models.UserToken) error {
	panic("implement me")
}

func (r *AuthRepository) GetUserToken(userID int) (*models.UserToken, error) {
	panic("implement me")
}

func (r *AuthRepository) RemoveUserTokens(userID int) (int, error) {
	panic("implement me")
}

func (r *AuthRepository) ClearUserTokens() error {
	panic("implement me")
}

func (r *AuthRepository) SetUserTokenInvalidByToken(token string) error {
	panic("implement me")
}

func (r *AuthRepository) SetUserTokenInvalidByUserID(userID int) error {
	panic("implement me")
}
