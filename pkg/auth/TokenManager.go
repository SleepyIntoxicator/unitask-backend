package auth

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"math/rand"
	"strconv"
	"time"
)

type UserToken struct {
	AccessToken  string
	RefreshToken string
	IssueTs      int64
	ExpiresTs    int64
	Revoked      bool
}

type ITokenManager interface {
	NewJWT(accessTTL, userID int) (string, error)
	GenerateRefreshToken() (string, error)
	ParseAccessToken(accessToken string) (string, error)
	NewAppToken(appT AppTokenData) (string, error)
	NewUserToken(uT UserTokenData) (UserToken, error)
}

type TokenManager struct {
	signingMethod   jwt.SigningMethod
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	rTokenLength int
	rTokenFormat []rune
	signingKey   string

	aTokenBlacklistManager AccessTokenBlacklistManager
}

type TokenManagerConfig struct {
	jwt.SigningMethod
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	RefreshTokenLength int
	RefreshTokenFormat []rune

	SigningKey string
}

type AppTokenData struct {
	AppSecret        string
	AppName          string
	AppID            string
	startTimestamp   time.Time
	expiresTimestamp time.Time
}

type UserTokenData struct {
	UserID uuid.UUID
}

type TokenClaims struct {
	jwt.StandardClaims
}

func NewTokenManager(config TokenManagerConfig) (*TokenManager, error) {
	err := validation.ValidateStruct(
		&config,
		validation.Field(&config.SigningKey, validation.Required, validation.NotNil),
		validation.Field(&config.SigningMethod, validation.Required, validation.NotNil),
		validation.Field(&config.RefreshTokenFormat, validation.Required, validation.RuneLength(16, 64)),
	)
	if err != nil {
		return nil, err
	}

	if config.RefreshTokenLength < 16 || config.RefreshTokenLength > 64 {
		return nil, fmt.Errorf("length of the refresh token must be in the range from 16 to 64")
	}

	return &TokenManager{
		signingKey:      config.SigningKey,
		signingMethod:   config.SigningMethod,
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
		rTokenLength:    config.RefreshTokenLength,
		rTokenFormat:    config.RefreshTokenFormat,
	}, nil
}

func (mgr *TokenManager) NewJWT(uT UserTokenData, timeNow time.Time) (string, error) {
	token := jwt.NewWithClaims(mgr.signingMethod, TokenClaims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  timeNow.Unix(),
			ExpiresAt: timeNow.Add(mgr.accessTokenTTL).Unix(),
			Subject:   uT.UserID.String(),
		},
	})

	signed, err := token.SignedString([]byte(mgr.signingKey))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (mgr *TokenManager) NewUserToken(uT UserTokenData) (UserToken, error) {
	timeNow := time.Now()

	aToken, err := mgr.NewJWT(uT, timeNow)
	if err != nil {
		return UserToken{}, err
	}
	rToken := GetRefreshToken(mgr.rTokenLength, mgr.rTokenFormat)

	userToken := UserToken{
		AccessToken:  aToken,
		RefreshToken: rToken,
		IssueTs:      timeNow.Unix(),
		ExpiresTs:    timeNow.Add(mgr.refreshTokenTTL).Unix(),
		Revoked:      false,
	}

	return userToken, nil
}

func (mgr *TokenManager) NewAppToken(appT AppTokenData) (string, error) {
	hashTime := appT.startTimestamp.Nanosecond() + appT.expiresTimestamp.Nanosecond()

	s1 := sha1.New()
	s1.Write([]byte(appT.AppSecret))
	s1.Write([]byte(strconv.Itoa(hashTime)))

	m := md5.New()
	m.Write(s1.Sum(nil))
	m.Write([]byte(appT.AppName))
	token := hex.EncodeToString(m.Sum(nil))

	return token, nil
}

func GetRefreshToken(n int, letters []rune) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
