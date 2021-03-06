package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"github.com/google/uuid"
	"time"
	//"backend/internal/store"
)

type AuthRepository struct {
	store *Store
}

//	App auth

func (r *AuthRepository) RegisterApp(app *models.RegisteredApp) error {
	query := `INSERT INTO registeredapp (id, app_name, app_secret) VALUES ($1, $2, $3) returning id`
	err := r.store.db.QueryRow(query, app.ID, app.AppName, app.AppSecret).Scan(&app.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthRepository) GetApp(appUUID uuid.UUID) (*models.RegisteredApp, error) {
	app := &models.RegisteredApp{}

	query := `SELECT id, app_name, app_secret FROM registeredapp WHERE id = $1`
	err := r.store.db.QueryRow(query, appUUID).Scan(
		&app.ID,
		&app.AppName,
		&app.AppSecret)
	return app, store.HandleErrorNoRows(err)
}

func (r *AuthRepository) DeleteApp(appUUID uuid.UUID) error {
	query := `DELETE FROM registeredapp WHERE id = $1`
	_, err := r.store.db.Exec(query, appUUID)
	return err
}

func (r *AuthRepository) AddAppToken(t *models.AppToken) error {
	query := `INSERT INTO apptoken 
    	(token, app_id, issue_timestamp, start_timestamp, expiration_timestamp) 
		VALUES($1, $2, $3, $4, $5) RETURNING token`
	return r.store.db.QueryRow(query,
		t.AppToken,
		t.AppID,
		t.IssueTokenTimestamp,
		t.StartTimestamp,
		t.ExpirationTimestamp).Scan(&t.AppToken)
}

func (r *AuthRepository) RemoveAppTokens(appUUID uuid.UUID) error {
	query := `DELETE FROM apptoken WHERE app_id = $1`
	_, err := r.store.db.Exec(query, appUUID)
	return err
}

func (r *AuthRepository) GetAppTokenInfo(token string) (*models.AppToken, error) {
	t := &models.AppToken{}

	query := `SELECT token, app_id, issue_timestamp, start_timestamp, expiration_timestamp FROM apptoken WHERE token = $1 limit 1`
	err := r.store.db.QueryRow(query, token).
		Scan(&t.AppToken,
			&t.AppID,
			&t.IssueTokenTimestamp,
			&t.StartTimestamp,
			&t.ExpirationTimestamp)
	return t, store.HandleErrorNoRows(err)
}

func (r *AuthRepository) GetAppTokenByAppUUID(appUUID uuid.UUID) (*models.AppToken, error) {
	t := &models.AppToken{}

	query := `SELECT token, app_id, issue_timestamp, start_timestamp, expiration_timestamp
				FROM apptoken WHERE app_id = $1 ORDER BY issue_timestamp DESC LIMIT 1`
	err := r.store.db.QueryRow(query, appUUID.String()).
		Scan(&t.AppToken,
			&t.AppID,
			&t.IssueTokenTimestamp,
			&t.StartTimestamp,
			&t.ExpirationTimestamp)
	return t, store.HandleErrorNoRows(err)
}

func (r *AuthRepository) GetAppByName(name string) (*models.RegisteredApp, error) {
	app := &models.RegisteredApp{}

	query := `SELECT id, app_name, app_secret FROM registeredapp WHERE app_name = $1`
	err := r.store.db.QueryRow(query, name).Scan(&app.ID, &app.AppName, &app.AppSecret)
	return app, store.HandleErrorNoRows(err)
}

//	User auth

func (r *AuthRepository) AddUserToken(t *models.UserToken) error {
	query := `DELETE FROM usertoken WHERE user_id = $1 AND
                            id NOT IN (SELECT id FROM usertoken WHERE user_id = $1
                                    	ORDER BY issue_token_timestamp DESC LIMIT 4)`
	_, err := r.store.db.Exec(query, t.UserID)
	if err != nil {
		return err
	}
	query = `INSERT INTO usertoken 
				(user_id, access_token, refresh_token, issue_token_timestamp, start_timestamp, expiration_timestamp)  
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.store.db.QueryRow(query,
		t.UserID,
		t.AccessToken,
		t.RefreshToken,
		t.IssueTokenTimestamp,
		t.StartTimestamp,
		t.ExpirationTimestamp,
	).Err()
}

func (r *AuthRepository) GetUserToken(userID int) (*models.UserToken, error) {
	u := &models.UserToken{}

	query := `SELECT id, user_id, access_token, refresh_token, issue_token_timestamp, start_timestamp, expiration_timestamp 
				FROM usertoken 
				WHERE user_id = $1 
				order by issue_token_timestamp desc limit 1`
	err := r.store.db.QueryRow(query, userID).Scan(
		&u.ID,
		&u.UserID,
		&u.AccessToken,
		&u.RefreshToken,
		&u.IssueTokenTimestamp,
		&u.StartTimestamp,
		&u.ExpirationTimestamp)
	return u, store.HandleErrorNoRows(err)
}

func (r *AuthRepository) AddUserRefreshToken(token *models.UserToken) error {
	query := `INSERT INTO usertoken VALUES ($1) RETURNING id`
	err := r.store.db.QueryRow(query, token).Scan(&token.ID)
	return err
}

func (r *AuthRepository) IsRefreshTokenExist(refreshToken string) (bool, error) {
	query := `SELECT FROM usertoken WHERE refresh_token = $1`
	err := r.store.db.QueryRow(query, refreshToken).Err()
	return store.HandleIsFieldFounded(err)
}

func (r *AuthRepository) RemoveUserTokens(userID int) (int, error) {
	deletedRows := 0
	query := `DELETE FROM usertoken 
				WHERE user_id = $1 AND id NOT IN 
					(SELECT id FROM usertoken WHERE user_id = $1
				    	ORDER BY issue_token_timestamp DESC 
				    	LIMIT 1)`
	err := r.store.db.QueryRow(query, userID).Scan(&deletedRows)
	return deletedRows, store.HandleIgnoreErrorNoRows(err)
}

func (r *AuthRepository) ClearUserTokens() error {
	query := `DELETE FROM usertoken WHERE
                            user_id in (select id from "user" where login = 'admin') AND
                            id NOT IN (SELECT id FROM usertoken WHERE user_id = (select id from "user" where login = 'admin')
                                    	ORDER BY issue_token_timestamp DESC LIMIT 1)`
	_, err := r.store.db.Exec(query)
	return err
}

func (r *AuthRepository) SetUserTokenInvalidByToken(token string) error {
	query := `UPDATE usertoken 
				SET	expiration_timestamp = $1, exit_timestamp = $1
				WHERE refresh_token = $2`
	_, err := r.store.db.Exec(query, time.Now(), token)
	return store.HandleErrorNoRows(err)
}

func (r *AuthRepository) SetUserTokenInvalidByUserID(userID int) error {
	query := `UPDATE usertoken
				SET expiration_timestamp = $1, exit_timestamp = $1
				WHERE user_id = $2 AND expiration_timestamp < $1`
	_, err := r.store.db.Exec(query, time.Now(), userID)
	return store.HandleErrorNoRows(err)
}
