package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

//	----	----	----	middleware handles	----	----	----

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(CtxKeyRequestID),
			"method":      r.Method,
			"request_uri": r.RequestURI,
		})
		logger.Infof("started")

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}

		logger.WithFields(logrus.Fields{
			"status_code":   rw.code,
			"status_text":   http.StatusText(rw.code),
			"response_time": time.Now().Sub(start),
		}).Infof("completed with code %d", rw.code)

		next.ServeHTTP(rw, r)
	})
}

func (s *server) authenticateApp(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appToken := r.Header.Get("X-App-Token")
		if appToken == "" {
			s.error(w, r, http.StatusUnauthorized, errAppTokenNotFound)
			return
		}

		isValid, err := s.services.Auth().IsAppTokenValid(appToken)
		if !isValid || err != nil {
			s.error(w, r, http.StatusUnauthorized, service.ErrInvalidAppToken)
			return
		}

		app, err := s.services.Auth().GetAppInfoByToken(appToken)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		logrus.WithFields(logrus.Fields{
			"request_id":  r.Context().Value(CtxKeyRequestID),
			"request_uri": r.RequestURI,
			"app_token":   appToken,
		}).Debug("the app authenticated successfuly")

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyAppID, app.ID)))
	})
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var accessToken string
		authorization := r.Header.Get("Authorization")
		if authorization == "" {
			s.respondHTML(w, r, http.StatusUnauthorized, errUnauthorizedHTML)
			return
		}
		_, err := fmt.Sscanf(authorization, "Bearer %s", &accessToken)
		if err != nil && err.Error() == "input does not match format" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("the authorization code has an incorrect format"))
			return
		} else if err != nil {
			s.errorV2(w, r, http.StatusBadRequest, ErrorInvalidAuthorizationKey)
			return
		}

		user, err := s.services.Auth().AuthenticateUser(accessToken)
		if err != nil {
			s.errorV2(w, r, http.StatusUnauthorized, models.New(err, http.StatusUnauthorized, "invalid_access_token"))
			return
		}

		if user.EncryptedPassword != "" || user.Password != "" {
			s.logger.Errorf(
				"FATAL vulnerability of sensetive data EncrPw: %s or Pw: %s",
				user.EncryptedPassword,
				user.Password)
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, user)))
	})
}

//	----	----	----	Handlers	----	----	----

func (s *server) handleAppRegister() http.HandlerFunc {
	type req struct {
		AppName string `json:"app_name"`
	}
	type res struct {
		AppID     string `json:"client_id"`
		AppName   string `json:"client_name"`
		AppSecret string `json:"client_secret"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &req{}

		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		if validation.IsEmpty(req.AppName) {
			s.error(w, r, http.StatusBadRequest, errors.New("the app name field can't be empty"))
			return
		}

		app := &models.RegisteredApp{
			AppName: req.AppName,
		}
		err = s.services.Auth().RegisterApp(app)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &res{
			AppID:     app.ID.String(),
			AppName:   app.AppName,
			AppSecret: app.AppSecret,
		}

		s.respond(w, r, http.StatusOK, res)
	}
}

//Авторизация приложения
func (s *server) handleAppAuthorization() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appToken := r.Header.Get("X-App-Token")
		//	Если приложение уже авторизовано
		if appToken != "" {
			isTokenValid, err := s.services.Auth().IsAppTokenValid(appToken)
			if err != nil && err != service.ErrInvalidAppToken {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			if isTokenValid {
				http.Redirect(w, r, "/api/v1/auth/app/delete", http.StatusFound)
			}
		}

		reqAppID := r.URL.Query().Get("client_id")
		appSecret := r.URL.Query().Get("client_secret")
		if "" == reqAppID {
			s.error(w, r, http.StatusBadRequest,
				errors.New(errInvalidRequest.Error()+": invalid client_id parameter"))
			return
		}
		if "" == appSecret {
			s.error(w, r, http.StatusBadRequest,
				errors.New(errInvalidRequest.Error()+": invalid client_secret parameter"))
			return
		}

		isAppSecretValid, err := s.services.Auth().IsAppSecretValid(reqAppID, appSecret)
		if err == service.ErrInvalidAppID || err == service.ErrAppAuthorization {
			s.error(w, r, http.StatusServiceUnavailable, err)
			return
		} else if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		if isAppSecretValid {
			appID, err := uuid.Parse(reqAppID)
			if err != nil {
				s.error(w, r, http.StatusUnauthorized, err)
				return
			}

			appToken, err := s.services.Auth().GetAppToken(appID)
			if err != nil {
				s.error(w, r, http.StatusServiceUnavailable, err)
				return
			}

			w.Header().Add("X-App-Token", appToken.AppToken)
			s.respond(w, r, http.StatusOK, appToken)
		}

	}
}

func (s *server) handleAppDelete() http.HandlerFunc {
	type request struct {
		AppID     string `json:"client_id"`
		AppSecret string `json:"client_secret"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err == errJSONEOF {
			s.error(w, r, http.StatusBadRequest, errMissingRequestBody)
			return
		} else if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		accessToken := r.Header.Get("X-App-Token")

		err = s.services.Auth().DeleteApp(req.AppID, req.AppSecret, accessToken)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		s.respond(w, r, http.StatusNoContent, nil)
	}
}

//Регистрация пользователя
func (s *server) handleUserRegister() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
	}
	type response struct {
		User models.User `json:"user"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			if err.Error() == errJSONEOF.Error() {
				s.error(w, r, http.StatusBadRequest, errMissingRequestBody)
				return
			}
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u := &models.User{
			Login:    req.Login,
			Email:    req.Email,
			FullName: req.FullName,
			Password: req.Password,
		}

		err := s.services.Auth().RegisterUser(u)
		if err != nil {
			if err != service.ErrEmailIsAlreadyOccupied || err != service.ErrLoginIsAlreadyOccupied {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			s.error(w, r, http.StatusOK, err)
			return
		}

		u.Sanitize()

		s.respond(w, r, http.StatusCreated, response{User: *u})
	}
}

func (s *server) handleUserSignIn() http.HandlerFunc {
	type signInRequest struct {
		Login    string `json:"login"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type userInfo struct {
		UserID    int       `json:"user_id"`
		FullName  string    `json:"full_name"`
		Login     string    `json:"login"`
		CreatedAt time.Time `json:"created_at"`
	}
	type response struct {
		User         userInfo `json:"user"`
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		Expires      int32    `json:"expires"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req := &signInRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		ip, port, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}
		s.logger.WithFields(logrus.Fields{
			"user_ip":   ip,
			"user_port": port,
		}).Debug("user was signed in")
		if ip == "" || port == "" {
			s.error(w, r, http.StatusBadRequest, fmt.Errorf("debuging: ip or port not found in request: %s:%s", ip, port))
			return
		}

		userToken, err := s.services.Auth().UserSignIn(&models.UserSignIn{
			Login:    req.Login,
			Email:    req.Email,
			Password: req.Password,
			IP:       ip,
		})
		if err != nil {
			s.error(w, r, http.StatusOK, err)
			return
		}

		user, err := s.services.User().Find(userToken.UserID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		user.Sanitize()

		res := &response{
			User: userInfo{
				UserID:    user.ID,
				FullName:  user.FullName,
				Login:     user.Login,
				CreatedAt: user.CreatedAt,
			},
			AccessToken:  userToken.AccessToken,
			RefreshToken: userToken.RefreshToken,
			Expires:      int32(userToken.ExpirationTimestamp.Sub(time.Now()).Seconds()),
		}
		s.respond(w, r, http.StatusOK, res)

	}
}

func (s *server) handleUserLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		err = s.services.Auth().UserLogout(user.ID)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) handleUserToken() http.HandlerFunc {
	type request struct {
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
	}
	type response struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var accessToken string
		_, err := fmt.Sscanf(r.Header.Get("Authorization"), "Bearer %s", &accessToken)
		if err != nil {
			s.errorV2(w, r, http.StatusOK, ErrorInvalidAuthorizationKey)
			return
		}

		req := &request{}
		err = json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			if err == errJSONEOF {
				s.error(w, r, http.StatusBadRequest, errJSONParseEOF)
				return
			}
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		user, err := s.services.Auth().AuthenticateUser(accessToken)
		switch err {
		case nil:
			//case error is nil
			break

		case service.ErrAccessTokenIsBlacklisted:
			s.error(w, r, http.StatusBadRequest, err)

			return
		default:
			// Manual user search, because the AuthUser() returned nil
			tokenClaims, err := s.services.Auth().CheckAccessToken(accessToken)
			if tokenClaims == nil {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			// Ignore error, if it's equals "Token expired"
			if err != nil && tokenClaims.VerifyExpiresAt(time.Now().Unix(), true) {
				s.error(w, r, http.StatusInternalServerError, err)
				return
			}
			user, err = s.services.User().Find(tokenClaims.UserID)
		}
		if user == nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		newToken, err := s.services.Auth().RefreshPairAccessRefreshToken(user.ID, accessToken, req.RefreshToken)
		switch err {
		case service.ErrInvalidTokenPair, service.ErrAccessTokenRefreshRateExceeded:
			s.error(w, r, http.StatusTooManyRequests, err)
			return
		case nil:
			break
		default:
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		res := &response{
			TokenType:    "bearer",
			AccessToken:  newToken.AccessToken,
			RefreshToken: newToken.RefreshToken,
			ExpiresIn:    newToken.ExpirationTimestamp.Unix(),
		}
		s.respond(w, r, http.StatusOK, res)
	}
}

/*
func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var accessToken string
		_, err := fmt.Sscanf(r.Header.Get("Authorization"), "Bearer %s", &accessToken)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		type claims struct {
			jwt.StandardClaims
			UserID int    `json:"user_id"`
			Login  string `json:"login"`
			Iat    int64  `json:"iat"`
			Exp    int64  `json:"exp"`
			Nbf    int64  `json:"nbf"`
		}

		tokenClaims := &claims{}

		token, err := jwt.ParseWithClaims(accessToken, &claims{}, func(*jwt.Token) (interface{}, error) {
			return []byte(s.services.Auth().GetSigningKey()), nil
		})
		if err != nil {
			s.error(w, r, http.StatusOK, err)
			return
		}
		tokenClaims, ok := token.Claims.(*claims)
		if !ok && !token.Valid {
			s.error(w, r, http.StatusOK, service.ErrInvalidUserToken)
			return
		}

		u, err := s.services.User().Find(tokenClaims.UserID)
		if err != nil {
			s.error(w, r, http.StatusOK, err)
			http.Redirect(w, r, "/api/v1/auth/login", http.StatusUnauthorized)
			return
		}

		if u == nil {
			fmt.Print("End point")
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, u)))
	})
}
*/
