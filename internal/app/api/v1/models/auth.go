package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type AppToken struct {
	AppID               uuid.UUID `json:"client_id" db:"app_id"`
	AppToken            string    `json:"access_token" db:"app_token"`
	IssueTokenTimestamp time.Time `json:"-" db:"issue_token_timestamp"`
	StartTimestamp      time.Time `json:"-" db:"start_timestamp"`
	ExpirationTimestamp time.Time `json:"expiration_timestamp" db:"expiration_timestamp"`
}

func (t *AppToken) Valid() bool {
	now := time.Now()
	if now.After(t.IssueTokenTimestamp) ||
		now.After(t.StartTimestamp) ||
		now.Before(t.ExpirationTimestamp) {
		return true
	}

	return false
}

func (t *AppToken) CheckIdentity(requestToken string) bool {
	if requestToken == t.AppToken {
		return true
	}
	return false
}

type RegisteredApp struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AppName   string    `json:"app_name" db:"app_name"`
	AppSecret string    `json:"app_secret" db:"app_secret"`
}

func (a *RegisteredApp) IsSecretValid(appSecret string) bool {
	if appSecret == a.AppSecret {
		return true
	}

	return false
}
//	----	----	----	----	----	----	----	----

type UserToken struct {
	ID                  int       `db:"id"`
	UserID              int       `db:"user_id"`
	AccessToken         string    `db:"access_token"`
	RefreshToken        string    `db:"refresh_token"`
	IssueTokenTimestamp time.Time `db:"issue_token_timestamp"`
	StartTimestamp      time.Time `db:"start_timestamp"`
	ExpirationTimestamp time.Time `db:"expiration_timestamp"`
	LogoutTimestamp     time.Time `db:"exit_timestamp"`
}

func (t *UserToken) Valid() bool {
	now := time.Now()
	if now.After(t.IssueTokenTimestamp) ||
		now.After(t.StartTimestamp) ||
		now.Before(t.ExpirationTimestamp) {
		return true
	}

	return false
}

type UAccessTokenClaims struct {
	UserID int    `json:"user_id"`
	//Exp    int64  `json:"exp"`
	jwt.StandardClaims
}

func (t *UAccessTokenClaims) Validate() error {
	now := time.Now().Unix()
	if now < t.IssuedAt {
		return ErrAccessTokenIsInvalid
	}
	if now > t.ExpiresAt {
		return ErrAccessTokenIsExpired
	}
	return nil
}