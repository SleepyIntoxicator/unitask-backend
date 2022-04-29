package apiserver

import (
	"backend/internal/api/v1/models"
	"backend/internal/config"
	"backend/internal/service"
	"backend/internal/service/services"
	"backend/internal/store"
	"backend/pkg/hooks"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

type ctxKey int8

const (
	CtxKeyUser = ctxKey(iota)
	CtxKeyRequestID
	CtxKeyAppID
)

type server struct {
	router     *mux.Router
	logger     *logrus.Logger
	services   *services.Service
	config     *config.Config
	logrusHook *hooks.LogsStoreHook
	remoteHook *hooks.LogsRemoteControllerHook
}

func newServer(store store.Store, config *config.Config) *server {
	s := &server{
		router:   mux.NewRouter(),
		logger:   logrus.New(),
		services: services.NewService(store, config),
		config:   config,
	}

	s.configureRouter()

	_, err := s.configureLogger(config.Logrus.Level)
	if err != nil {
		s.logger.Error("cannot configure router: ", err)
	}

	s.services.AddLogger(s.logger)

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.configureUnauthorizedHandlers()

	s.configureRouterAPIv1()
}

func (s *server) configureLogger(loggerLevel string) (*logrus.Logger, error) {
	level, err := logrus.ParseLevel(loggerLevel)
	if err != nil {
		return nil, err
	}

	formatter := &logrus.TextFormatter{
		DisableSorting:   false, //false
		DisableTimestamp: false, //false
		ForceColors:      true,  //true
		PadLevelText:     false, //false
	}

	s.logrusHook = new(hooks.LogsStoreHook)
	//s.remoteHook = hooks.NewLogRemoteController(...)

	s.logger.SetFormatter(formatter)
	s.logger.SetOutput(os.Stdout)
	s.logger.SetLevel(level)
	s.logger.Hooks.Add(s.logrusHook)

	logrus.SetFormatter(formatter)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(level)
	logrus.AddHook(s.logrusHook)

	return s.logger, nil
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	requestID, internalError := s.getRequestIDFromContext(r.Context())
	if internalError != nil {
		s.logger.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote_addr":    r.RemoteAddr,
			"request_id":     requestID,
			"request_uri":    r.RequestURI,
			"internal_error": internalError,
		}).Error("requestID not found in context")
	}

	s.logger.WithFields(logrus.Fields{
		"method":          r.Method,
		"remote_addr":     r.RemoteAddr,
		"request_id":      requestID,
		"request_uri":     r.RequestURI,
		"request_z_error": err,
	}).Error("completed with error")
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) errorV2(w http.ResponseWriter, r *http.Request, code int, err models.ServerError) {
	requestID, internalError := s.getRequestIDFromContext(r.Context())
	if internalError != nil {
		s.logger.WithFields(logrus.Fields{
			"method":         r.Method,
			"remote_addr":    r.RemoteAddr,
			"request_id":     requestID,
			"request_uri":    r.RequestURI,
			"internal_error": internalError,
		}).Error("requestID not found in context")
	}

	s.logger.WithFields(logrus.Fields{
		"method":          r.Method,
		"remote_addr":     r.RemoteAddr,
		"request_id":      requestID,
		"request_uri":     r.RequestURI,
		"request_z_error": err,
	}).Error("completed with error")
	s.respond(w, r, 200, map[string]models.ServerError{"error": err})
}

func (s *server) respondErrors(w http.ResponseWriter, r *http.Request, code int, errs []error) {
	type Error struct {
		Code         int    `json:"code"`
		ErrorMessage string `json:"error_message"`
		ErrorCode    string `json:"error_description"`
	}

	type errorsSlice struct {
		Errors []Error `json:"errors"`
	}
	respondErrors := errorsSlice{}
	for i := range errs {
		err := Error{
			Code:         code,
			ErrorMessage: errs[i].Error(),
		}
		respondErrors.Errors = append(respondErrors.Errors, err)
	}

	jsonError := json.NewEncoder(w).Encode(respondErrors)
	if jsonError != nil {
		s.logger.WithFields(logrus.Fields{
			"errors": respondErrors,
		}).Error("Cannot encode respond errors")
	}
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			_ = json.NewEncoder(w).Encode(err)
		}
	}
}

func (s *server) respondHTML(w http.ResponseWriter, r *http.Request, code int, data string) {
	if code >= 400 {
		requestID, internalError := s.getRequestIDFromContext(r.Context())
		if internalError != nil {
			s.logger.WithFields(logrus.Fields{
				"method":         r.Method,
				"remote_addr":    r.RemoteAddr,
				"request_id":     requestID,
				"request_uri":    r.RequestURI,
				"internal_error": internalError,
			}).Error("requestID not found in context")
		}

		s.logger.WithFields(logrus.Fields{
			"method":          r.Method,
			"remote_addr":     r.RemoteAddr,
			"request_id":      requestID,
			"request_uri":     r.RequestURI,
			"request_z_error": data,
		}).Error("completed with error")
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(data))
}

func (s *server) getLimitAndOffsetFromQuery(r *http.Request) (limit int, offset int, err error) {
	lim := r.URL.Query().Get("limit")
	of := r.URL.Query().Get("offset")

	if len(lim) > 8 || len(of) > 8 {
		return 0, 0, models.ErrLimitOrOffsetTooLarge
	}

	if lim == "" {
		limit = 0
	} else {

		limit, err = strconv.Atoi(lim)
		if err != nil {
			return 0, 0, err
		}
		if limit < 0 {
			return 0, 0, models.ErrLimitLessThanZero
		}
	}
	if of == "" {
		offset = 0
	} else {
		offset, err = strconv.Atoi(of)
		if err != nil {
			return 0, 0, err
		} else if offset < 0 {
			return 0, 0, models.ErrOffsetLessThanZero
		}
	}
	return
}

func (s *server) getUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(CtxKeyUser).(*models.User)
	if !ok {
		return nil, errUserNotFoundInContext
	}
	return user, nil
}

func (s *server) getRequestIDFromContext(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(CtxKeyRequestID).(uuid.UUID)
	if !ok {
		return uid, service.ErrRequestIDNotFoundInContext
	}

	return uid, nil
}

func (s *server) getAppIDFromContext(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(CtxKeyAppID).(uuid.UUID)
	if !ok {
		return uid, service.ErrAppIDNotFoundInContext
	}

	return uid, nil
}
