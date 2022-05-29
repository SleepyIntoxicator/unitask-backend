package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"strconv"
	"time"
)

const (
	refreshTokenLength = 64
)

const (
	minimumAccessTokenRefreshRate = 5 * time.Second
)

type tokenClaims struct {
	jwt.StandardClaims
	UserID int `json:"user_id"`
}

type AuthService struct {
	service *Service

	tokenBlacklistMgr *BlacklistManager
	signingKey        string
	accessTTL         time.Duration
	refreshTTL        time.Duration
}

func NewAuthService(service *Service) *AuthService {
	s := &AuthService{
		service:    service,
		refreshTTL: service.config.Auth.JWT.RefreshTokenTTL,
		accessTTL:  service.config.Auth.JWT.AccessTokenTTL,
	}

	s.signingKey = service.config.Auth.JWT.SigningKey

	s.tokenBlacklistMgr = New()
	_ = service.store.Auth().ClearUserTokens(context.Background())
	return s
}

func (s *AuthService) RegisterApp(ctx context.Context, app *models.RegisteredApp) error {
	_, err := s.service.store.Auth().GetAppByName(app.AppName)
	if err != nil && err != store.ErrRecordNotFound {
		return err
	} else if err == nil {
		return service.ErrAppNameIsAlreadyOccupied
	}

	appUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	app.ID = appUUID
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	app.AppSecret = s.getToken(32, letters)
	err = s.service.store.Auth().RegisterApp(ctx, app)

	return err
}

func (s *AuthService) DeleteApp(ctx context.Context, appID, appSecret, appToken string) error {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return service.ErrInvalidAppID
	}

	regApp, err := s.service.store.Auth().GetApp(ctx, appUUID)
	if err != nil {
		return service.ErrAppNotFound
	}

	regAppToken, err := s.service.store.Auth().GetAppTokenByAppUUID(ctx, appUUID)
	if err != nil {
		return service.ErrAppAuthorization
	}

	if regApp.AppSecret != appSecret || regAppToken.AppToken != appToken {
		return service.ErrAppAuthorization
	}

	err = s.service.store.Auth().RemoveAppTokens(ctx, appUUID)
	if err != nil {
		return err
	}

	err = s.service.store.Auth().DeleteApp(ctx, appUUID)
	if err != nil {
		return err
	}

	return nil
}

//RegisterUser Returns service.ErrEmailIsAlreadyOccupied or service.ErrLoginIsAlreadyOccupied
//if a user with this login or email already exist
func (s *AuthService) RegisterUser(ctx context.Context, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	_, err := s.service.User().FindByEmail(ctx, user.Email)
	if err != nil && err != service.ErrUserNotFound {
		return err
	} else if err != service.ErrUserNotFound {
		return service.ErrEmailIsAlreadyOccupied
	}

	_, err = s.service.User().FindByLogin(ctx, user.Login)
	if err != nil && err != service.ErrUserNotFound {
		return err
	} else if err != service.ErrUserNotFound {
		return service.ErrLoginIsAlreadyOccupied
	}

	if err := user.BeforeCreate(); err != nil {
		return err
	}

	err = s.service.store.User().Create(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) UserSignIn(ctx context.Context, userSignIn *models.UserSignIn) (*models.UserToken, error) {
	user := &models.User{}
	var err error

	if userSignIn.Login != "" && userSignIn.Email == "" {
		user, err = s.service.store.User().FindByLogin(ctx, userSignIn.Login)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrIncorrectLoginOrPassword
		} else if err != nil {
			return nil, err
		}
	} else if userSignIn.Email != "" && userSignIn.Login == "" {

		user, err = s.service.store.User().FindByEmail(ctx, userSignIn.Email)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrIncorrectLoginOrPassword
		} else if err != nil {
			return nil, err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte(userSignIn.Password)); err != nil {
		return nil, service.ErrIncorrectLoginOrPassword
	}

	accessToken, err := s.service.Auth().GenerateAccessToken(user, time.Now())
	if err != nil {
		return nil, err
	}
	refreshToken := s.service.Auth().GenerateRefreshToken()
	userToken := &models.UserToken{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		IssueTokenTimestamp: time.Now(),
		ExpirationTimestamp: time.Now().Add(s.refreshTTL),
		StartTimestamp:      time.Now(),
		UserID:              user.ID,
	}

	if err := s.service.store.Auth().AddUserToken(ctx, userToken); err != nil {
		return nil, err
	}

	return userToken, nil
}

func (s *AuthService) UserLogout(ctx context.Context, userID int) error {
	//TODO Выходит сразу со всех устройств и приложений. Доработать
	token, err := s.service.store.Auth().GetUserToken(ctx, userID)
	if err != nil {
		return err
	}

	s.addAccessTokenToBlacklist(token.AccessToken, token.ExpirationTimestamp)

	err = s.service.store.Auth().SetUserTokenInvalidByToken(ctx, token.AccessToken)
	if err != nil {
		return err
	}

	_, err = s.service.store.Auth().RemoveUserTokens(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) addAccessTokenToBlacklist(token string, expirationTimestamp time.Time) {
	s.tokenBlacklistMgr.AddTokenToBlacklist(token, expirationTimestamp)
}

func (s *AuthService) AuthenticateUser(ctx context.Context, accessToken string) (*models.User, error) {
	tokenClaims, err := s.CheckAccessToken(accessToken)
	if err != nil {
		return nil, err
	}

	if err := tokenClaims.Valid(); err != nil {
		return nil, err
	}

	if s.tokenBlacklistMgr.IsTokenBlacklisted(accessToken) {
		return nil, service.ErrAccessTokenIsBlacklisted
	}

	u, err := s.service.User().Find(ctx, tokenClaims.UserID)
	if err != nil {
		return nil, err
	}

	return u, nil

}

func (s *AuthService) RefreshPairAccessRefreshToken(ctx context.Context, userID int, accessToken, refreshToken string) (*models.UserToken, error) {
	// Проверить наличие и валидность refresh-токена
	oldToken, err := s.service.Auth().CheckRefreshToken(ctx, refreshToken, userID)
	if err != nil {
		return nil, err
	}

	// Проверить соответствие пары токенов
	if oldToken.AccessToken != accessToken {
		return nil, service.ErrInvalidTokenPair
	}

	// Получить данные токенов
	oldTokenClaims, err := s.CheckAccessToken(accessToken)
	if oldTokenClaims == nil {
		return nil, err
	}
	if err != nil && oldTokenClaims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, err
	}

	// Выдать новую пару токенов только после истечения предыдущего
	LastUpdateTime := time.Unix(oldTokenClaims.IssuedAt, 0).Add(minimumAccessTokenRefreshRate)
	if time.Now().Before(LastUpdateTime) {
		return nil, service.ErrAccessTokenRefreshRateExceeded
	}

	// Выдать время окончания жизни токена сейчас
	if err := s.service.store.Auth().SetUserTokenInvalidByToken(ctx, oldToken.RefreshToken); err != nil {
		return nil, err
	}

	// Получить пользователя для генерации нового токена
	user, err := s.service.store.User().Find(ctx, userID)
	if err != nil {
		return nil, err
	}

	//Выдать новую пару access-refresh
	newToken, err := s.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// Добавить новую пару токенов в бд
	if err := s.service.store.Auth().AddUserToken(ctx, newToken); err != nil {
		return nil, err
	}

	return newToken, nil
}

func (s *AuthService) CheckAccessToken(accessToken string) (*models.UAccessTokenClaims, error) {
	tokenClaims := &models.UAccessTokenClaims{}

	token, err := jwt.ParseWithClaims(accessToken, &models.UAccessTokenClaims{}, func(*jwt.Token) (interface{}, error) {
		return []byte(s.signingKey), nil
	})
	if token == nil {
		return nil, err
	}

	tokenClaims, ok := token.Claims.(*models.UAccessTokenClaims)
	isExpired := !tokenClaims.VerifyExpiresAt(time.Now().Unix(), true)
	if !ok && token.Valid && !isExpired {
		return nil, service.ErrInvalidUserToken
	}
	return tokenClaims, err
}

func (s *AuthService) CheckRefreshToken(ctx context.Context, refreshToken string, userID int) (*models.UserToken, error) {
	userToken, err := s.service.store.Auth().GetUserToken(ctx, userID)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, service.ErrInvalidUserToken
	} else if err != nil {
		return nil, err
	}

	//Проверка на соответствие refresh токена
	if userToken.RefreshToken != refreshToken {
		return nil, service.ErrInvalidRefreshToken
	}

	//	Проверить валидность токена
	if !userToken.Valid() {
		return nil, service.ErrInvalidUserToken
	}

	return userToken, nil
}

func (s *AuthService) IsAppSecretValid(ctx context.Context, appID, appSecret string) (bool, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return false, service.ErrAppAuthorization
	}

	appData, err := s.service.store.Auth().GetApp(ctx, appUUID)
	if err != nil {
		if err == store.ErrNoRowsFound {
			return false, service.ErrInvalidAppID
		}
		return false, err
	}

	if !appData.IsSecretValid(appSecret) {
		return false, service.ErrAppAuthorization
	}

	return true, nil
}

func (s *AuthService) IsAppTokenValid(ctx context.Context, appToken string) (bool, error) {
	storeToken, err := s.service.store.Auth().GetAppTokenInfo(ctx, appToken)
	if err == store.ErrRecordNotFound {
		return false, service.ErrInvalidAppToken
	} else if err != nil {
		return false, err
	}

	if !storeToken.CheckIdentity(appToken) {
		return false, service.ErrInvalidAppToken
	}

	if !storeToken.Valid() {
		return false, service.ErrInvalidAppToken
	}

	return true, nil
}

func (s *AuthService) GetAppToken(ctx context.Context, appID uuid.UUID) (*models.AppToken, error) {
	token, err := s.service.store.Auth().GetAppTokenByAppUUID(ctx, appID)
	//Если необработанная ошибка
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
		//Если токен не существует
	} else if err == store.ErrRecordNotFound || (err == nil && time.Now().After(token.ExpirationTimestamp)) {
		//Генерируем новый
		app, err := s.service.store.Auth().GetApp(ctx, appID)
		if err != nil {
			return nil, err
		}

		//Добавляем новый токен в базу
		newAppToken := &models.AppToken{
			IssueTokenTimestamp: time.Now(),
			ExpirationTimestamp: time.Now().Add(s.service.config.Auth.AppTokenTTL),
			StartTimestamp:      time.Now(),
			AppID:               appID,
		}

		newAppToken.AppToken, err = s.GenerateAppToken(app, newAppToken.StartTimestamp, newAppToken.ExpirationTimestamp)
		if err != nil {
			return nil, err
		}

		if time.Now().After(token.ExpirationTimestamp) {
			err = s.service.store.Auth().RemoveAppTokens(ctx, token.AppID)
			if err != nil {
				return nil, err
			}
		}
		if err := s.service.store.Auth().AddAppToken(ctx, newAppToken); err != nil {
			return nil, err
		}

		return newAppToken, nil
	}
	return token, nil
}

func (s *AuthService) GetAppInfoByAppToken(ctx context.Context, token string) (*models.RegisteredApp, error) {
	tokenInfo, err := s.GetAppTokenInfo(ctx, token)
	if err != nil {
		return nil, err
	}

	app, err := s.service.store.Auth().GetApp(ctx, tokenInfo.AppID)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrInvalidAppID
	} else if err != nil {
		return nil, err
	}
	return app, nil
}

// GetAppTokenInfo returns the app token from store
func (s *AuthService) GetAppTokenInfo(ctx context.Context, token string) (*models.AppToken, error) {
	// Load token from db
	tokenInfo, err := s.service.store.Auth().GetAppTokenInfo(ctx, token)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrInvalidAppToken
	} else if err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *AuthService) GenerateTokenPair(user *models.User) (*models.UserToken, error) {
	accessToken, err := s.service.Auth().GenerateAccessToken(user, time.Now())
	if err != nil {
		return nil, err
	}
	refreshToken := s.service.Auth().GenerateRefreshToken()
	token := &models.UserToken{
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		IssueTokenTimestamp: time.Now(),
		ExpirationTimestamp: time.Now().Add(s.refreshTTL),
		StartTimestamp:      time.Now(),
		UserID:              user.ID,
	}
	return token, nil
}

func (s *AuthService) GenerateAppToken(app *models.RegisteredApp, startTimestamp time.Time, expirationTimestamp time.Time) (string, error) {
	hashTime := startTimestamp.Nanosecond() + expirationTimestamp.Nanosecond()

	s1 := sha1.New()
	s1.Write([]byte(app.AppSecret))
	s1.Write([]byte(strconv.Itoa(hashTime)))

	m := md5.New()
	m.Write(s1.Sum(nil))
	m.Write([]byte(app.AppName))
	token := hex.EncodeToString(m.Sum(nil))

	return token, nil
}

func (s *AuthService) GenerateAccessToken(user *models.User, startTimestamp time.Time) (string, error) {
	jwtClaims := models.UAccessTokenClaims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(s.accessTTL).Unix(),
			Subject:   user.Login,
		},
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	signed, err := jwtToken.SignedString([]byte(s.signingKey))
	if err != nil {
		return signed, err
	}

	return signed, nil
}

// GenerateRefreshToken Generates the refresh token. Token is [a-zA-Z0-9]
func (s *AuthService) GenerateRefreshToken() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	return s.getToken(refreshTokenLength, letters)
}

func (s *AuthService) getToken(n int, letters []rune) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
