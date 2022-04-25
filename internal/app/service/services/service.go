package services

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/config"
	"back-end/internal/app/service"
	"back-end/internal/app/store"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ctxKey int8

const (
	CtxKeyUser = ctxKey(iota)
	CtxKeyRequestID
	CtxKeyAppID
)

type Service struct {
	store  store.Store
	logger *logrus.Logger
	config *config.Config

	authService       *AuthService
	userService       *UserService
	taskService       *TaskService
	universityService *UniversityService
	groupService      *GroupService
	subjectService    *SubjectService
}

func NewService(store store.Store, config *config.Config) *Service {
	return &Service{
		store:  store,
		config: config,
	}
}

func (s *Service) AddLogger(logger *logrus.Logger) {
	s.logger = logger
}

func (s *Service) Auth() service.AuthService {
	if s.authService == nil {
		s.authService = NewAuthService(s)
		s.logger.Info("The auth service was started")
	}

	return s.authService
}

func (s *Service) User() service.UserService {
	if s.userService == nil {
		s.userService = &UserService{
			service: s,
		}
		s.logger.Info("The user service was started")
	}

	return s.userService
}

func (s *Service) Task() service.TaskService {
	if s.taskService == nil {
		s.taskService = &TaskService{
			service: s,
		}
		s.logger.Info("The task service was started")
	}

	return s.taskService
}

func (s *Service) University() service.UniversityService {
	if s.universityService == nil {
		s.universityService = &UniversityService{
			service: s,
		}
		s.logger.Info("The university service was started")
	}
	return s.universityService
}

func (s *Service) Group() service.GroupService {
	if s.groupService == nil {
		s.groupService = &GroupService{
			service: s,
		}
		s.logger.Info("The group service was started")
	}

	return s.groupService
}

func (s *Service) Subject() service.SubjectService {
	if s.subjectService == nil {
		s.subjectService = &SubjectService{
			service: s,
		}
		s.logger.Info("The subject service was started")
	}

	return s.subjectService
}

func (s *Service) getUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(CtxKeyUser).(*models.User)
	if !ok {
		return nil, service.ErrUserNotFoundInContext
	}

	return user, nil
}

func (s *Service) getRequestIDFromContext(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(CtxKeyRequestID).(uuid.UUID)
	if !ok {
		return uid, service.ErrRequestIDNotFoundInContext
	}

	return uid, nil
}

func (s *Service) getAppIDFromContext(ctx context.Context) (uuid.UUID, error) {
	uid, ok := ctx.Value(CtxKeyAppID).(uuid.UUID)
	if !ok {
		return uid, service.ErrAppIDNotFoundInContext
	}

	return uid, nil
}
