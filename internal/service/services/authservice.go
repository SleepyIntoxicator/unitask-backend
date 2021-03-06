package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
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

var (
	TokenTTL         = 12 * time.Hour      //Deprecated.
	appTokenTTL      = 24 * time.Hour * 30 //Deprecated.
	signingKeyLength = 64                  //Deprecated.

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
	_ = service.store.Auth().ClearUserTokens()
	return s
}

func (s *AuthService) GetSigningKey() string {
	return s.signingKey
}

func (s *AuthService) RegisterApp(app *models.RegisteredApp) error {
	_, err := s.service.store.Auth().GetAppByName(app.AppName)
	if err != nil && err != store.ErrRecordNotFound {
		return err
	} else if err == nil {
		return service.ErrAppNameIsAlreadyOccupied
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	app.ID = uuid
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	app.AppSecret = s.GetToken(32, letters)
	err = s.service.store.Auth().RegisterApp(app)

	return err
}

func (s *AuthService) DeleteApp(appID, appSecret, appToken string) error {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return service.ErrInvalidAppID
	}

	regApp, err := s.service.store.Auth().GetApp(appUUID)
	if err != nil {
		return service.ErrAppNotFound
	}

	regAppToken, err := s.service.store.Auth().GetAppTokenByAppUUID(appUUID)
	if err != nil {
		return service.ErrAppAuthorization
	}

	if regApp.AppSecret != appSecret || regAppToken.AppToken != appToken {
		return service.ErrAppAuthorization
	}

	err = s.service.store.Auth().RemoveAppTokens(appUUID)
	if err != nil {
		return err
	}

	err = s.service.store.Auth().DeleteApp(appUUID)
	if err != nil {
		return err
	}

	return nil
}

//RegisterUser Returns service.ErrEmailIsAlreadyOccupied or service.ErrLoginIsAlreadyOccupied
//if a user with this login or email already exist
func (s *AuthService) RegisterUser(user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	_, err := s.service.User().FindByEmail(user.Email)
	if err != nil && err != service.ErrUserNotFound {
		return err
	} else if err != service.ErrUserNotFound {
		return service.ErrEmailIsAlreadyOccupied
	}

	_, err = s.service.User().FindByLogin(user.Login)
	if err != nil && err != service.ErrUserNotFound {
		return err
	} else if err != service.ErrUserNotFound {
		return service.ErrLoginIsAlreadyOccupied
	}

	if err := user.BeforeCreate(); err != nil {
		return err
	}

	err = s.service.store.User().Create(user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) UserSignIn(userSignIn *models.UserSignIn) (*models.UserToken, error) {
	user := &models.User{}
	var err error

	if userSignIn.Login != "" && userSignIn.Email == "" {
		user, err = s.service.store.User().FindByLogin(userSignIn.Login)
		if err == store.ErrRecordNotFound {
			return nil, service.ErrIncorrectLoginOrPassword
		} else if err != nil {
			return nil, err
		}
	} else if userSignIn.Email != "" && userSignIn.Login == "" {

		user, err = s.service.store.User().FindByEmail(userSignIn.Email)
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

	if err := s.service.store.Auth().AddUserToken(userToken); err != nil {
		return nil, err
	}

	return userToken, nil
}

func (s *AuthService) UserLogout(userID int) error {
	//TODO ?????????????? ?????????? ???? ???????? ?????????????????? ?? ????????????????????. ????????????????????
	token, err := s.service.store.Auth().GetUserToken(userID)
	if err != nil {
		return err
	}

	s.addAccessTokenToBlacklist(token.AccessToken, token.ExpirationTimestamp)

	err = s.service.store.Auth().SetUserTokenInvalidByToken(token.AccessToken)
	if err != nil {
		return err
	}

	_, err = s.service.store.Auth().RemoveUserTokens(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) AuthenticateUser(accessToken string) (*models.User, error) {
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

	u, err := s.service.User().Find(tokenClaims.UserID)
	if err != nil {
		return nil, err
	}

	return u, nil

}

func (s *AuthService) RefreshPairAccessRefreshToken(userID int, accessToken, refreshToken string) (*models.UserToken, error) {
	// ?????????????????? ?????????????? ?? ???????????????????? refresh-????????????
	oldToken, err := s.CheckRefreshToken(refreshToken, userID)
	if err != nil {
		return nil, err
	}

	// ?????????????????? ???????????????????????? ???????? ??????????????
	if oldToken.AccessToken != accessToken {
		return nil, service.ErrInvalidTokenPair
	}

	// ???????????????? ???????????? ??????????????
	oldTokenClaims, err := s.CheckAccessToken(accessToken)
	if oldTokenClaims == nil {
		return nil, err
	}
	if err != nil && oldTokenClaims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, err
	}

	// ???????????? ?????????? ???????? ?????????????? ???????????? ?????????? ?????????????????? ??????????????????????
	LastUpdateTime := time.Unix(oldTokenClaims.IssuedAt, 0).Add(minimumAccessTokenRefreshRate)
	if time.Now().Before(LastUpdateTime) {
		return nil, service.ErrAccessTokenRefreshRateExceeded
	}

	// ???????????? ?????????? ?????????????????? ?????????? ???????????? ????????????
	if err := s.service.store.Auth().SetUserTokenInvalidByToken(oldToken.RefreshToken); err != nil {
		return nil, err
	}

	// ???????????????? ???????????????????????? ?????? ?????????????????? ???????????? ????????????
	user, err := s.service.store.User().Find(userID)
	if err != nil {
		return nil, err
	}

	//???????????? ?????????? ???????? access-refresh
	newToken, err := s.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// ???????????????? ?????????? ???????? ?????????????? ?? ????
	if err := s.service.store.Auth().AddUserToken(newToken); err != nil {
		return nil, err
	}

	return newToken, nil
}

// GenerateJWTToken doesn't work correctly. Deprecated.
func (s *AuthService) GenerateJWTToken(userID int) (*jwt.Token, error) {
	user, err := s.service.User().Find(userID)
	if err != nil {
		return nil, err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES256, &tokenClaims{
		jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(TokenTTL).Unix(),
		},
		user.ID,
	})

	return t, errors.New("to do me")
}

func (s *AuthService) RestoreJWTToken(userID int, refreshToken string) (*models.UserToken, error) {
	panic("implement me")
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

func (s *AuthService) CheckRefreshToken(refreshToken string, userID int) (*models.UserToken, error) {
	userToken, err := s.service.store.Auth().GetUserToken(userID)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, service.ErrInvalidUserToken
	} else if err != nil {
		return nil, err
	}

	//???????????????? ???? ???????????????????????? refresh ????????????
	if userToken.RefreshToken != refreshToken {
		return nil, service.ErrInvalidRefreshToken
	}

	//	?????????????????? ???????????????????? ????????????
	if !userToken.Valid() {
		return nil, service.ErrInvalidUserToken
	}

	return userToken, nil
}

func (s *AuthService) IsAppSecretValid(appID, appSecret string) (bool, error) {
	appUUID, err := uuid.Parse(appID)
	if err != nil {
		return false, service.ErrAppAuthorization
	}

	appData, err := s.service.store.Auth().GetApp(appUUID)
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

func (s *AuthService) IsAppTokenValid(appToken string) (bool, error) {
	storeToken, err := s.service.store.Auth().GetAppTokenInfo(appToken)
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

func (s *AuthService) GetAppToken(appID uuid.UUID) (*models.AppToken, error) {
	token, err := s.service.store.Auth().GetAppTokenByAppUUID(appID)
	//???????? ???????????????????????????? ????????????
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
		//???????? ?????????? ???? ????????????????????
	} else if err == store.ErrRecordNotFound || (err == nil && time.Now().After(token.ExpirationTimestamp)) {
		//???????????????????? ??????????
		app, err := s.service.store.Auth().GetApp(appID)
		if err != nil {
			return nil, err
		}

		//?????????????????? ?????????? ?????????? ?? ????????
		newAppToken := &models.AppToken{
			IssueTokenTimestamp: time.Now(),
			ExpirationTimestamp: time.Now().Add(appTokenTTL),
			StartTimestamp:      time.Now(),
			AppID:               appID,
		}

		newAppToken.AppToken, err = s.GenerateAppToken(app, newAppToken.StartTimestamp, newAppToken.ExpirationTimestamp)
		if err != nil {
			return nil, err
		}

		if time.Now().After(token.ExpirationTimestamp) {
			err = s.service.store.Auth().RemoveAppTokens(token.AppID)
			if err != nil {
				return nil, err
			}
		}
		if err := s.service.store.Auth().AddAppToken(newAppToken); err != nil {
			return nil, err
		}

		return newAppToken, nil
	}
	return token, nil
}

func (s *AuthService) GetAppInfoByToken(token string) (*models.RegisteredApp, error) {
	tokenInfo, err := s.GetTokenInfo(token)
	if err != nil {
		return nil, err
	}

	app, err := s.service.store.Auth().GetApp(tokenInfo.AppID)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrInvalidAppID
	} else if err != nil {
		return nil, err
	}
	return app, nil
}

func (s *AuthService) GetTokenInfo(token string) (*models.AppToken, error) {
	tokenInfo, err := s.service.store.Auth().GetAppTokenInfo(token)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrInvalidAppToken
	} else if err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *AuthService) addAccessTokenToBlacklist(token string, expirationTimestamp time.Time) {
	s.tokenBlacklistMgr.AddTokenToBlacklist(token, expirationTimestamp)
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
	return s.GetToken(refreshTokenLength, letters)
}

// changeSigningKey. Deprecated.
// Generates the signing key. Key is [a-zA-Z0-9!-}...].
func (s *AuthService) changeSigningKey() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	s.signingKey = s.GetToken(signingKeyLength, letters)
	return s.signingKey
}

func (s *AuthService) GetToken(n int, letters []rune) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

//b := make([]byte, n)
//rBytes := []byte(string(letters[rand.Intn(len(letters))]))
//b[i] = rBytes[rand.Intn(len(rBytes))]
