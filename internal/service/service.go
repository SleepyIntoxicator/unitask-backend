package service

import (
	"backend/internal/api/v1/models"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type Service interface {
	Auth() AuthService
	User() UserService
	Task() TaskService
	University() UniversityService
	Group() GroupService
	Subject() SubjectService

	AddLogger(logger *logrus.Logger)
}

type AuthService interface {
	RegisterApp(ctx context.Context, app *models.RegisteredApp) error

	DeleteApp(ctx context.Context, appID, appSecret, appToken string) error

	// RegisterUser Register new user. Receive *model.User. Return ErrMailLoginAlreadyUsing or nil
	RegisterUser(ctx context.Context, user *models.User) error

	UserSignIn(ctx context.Context, userSignIn *models.UserSignIn) (*models.UserToken, error)

	UserLogout(ctx context.Context, userID int) error

	AuthenticateUser(ctx context.Context, accessToken string) (*models.User, error)

	RefreshPairAccessRefreshToken(ctx context.Context, userID int, accessToken, refreshToken string) (*models.UserToken, error)

	// CheckAccessToken Check accessToken and returns valid=true, if it is valid.
	CheckAccessToken(accessToken string) (*models.UAccessTokenClaims, error)

	// CheckRefreshToken Check refreshToken and returns valid=true, if it is valid.
	CheckRefreshToken(ctx context.Context, refreshToken string, userID int) (*models.UserToken, error)

	// IsAppSecretValid App secret validation. Returns true, if it's valid.
	//May returns service.ErrNoRowsFound or else.
	IsAppSecretValid(ctx context.Context, appID, appSecret string) (bool, error)

	// IsAppTokenValid App token validation. Returns true, if valid.
	IsAppTokenValid(ctx context.Context, appToken string) (bool, error)

	GetAppToken(ctx context.Context, appID uuid.UUID) (*models.AppToken, error)

	//GetAppInfoByAppToken ...
	GetAppInfoByAppToken(ctx context.Context, token string) (*models.RegisteredApp, error)

	//GetAppTokenInfo returns the app token from store.
	GetAppTokenInfo(ctx context.Context, token string) (*models.AppToken, error)

	GenerateAppToken(app *models.RegisteredApp, startTimestamp time.Time, expirationTimestamp time.Time) (string, error)

	GenerateAccessToken(user *models.User, startTimestamp time.Time) (string, error)

	GenerateRefreshToken() string
}

type UserService interface {
	Create(*models.User) error
	GetAllUsers(limit, offset int) ([]models.User, error)

	// Find returns *models.User by userID or nil if user not found.
	//	If user was not found, method returns service.ErrUserNotFound.
	//	If an error occurs during the execution of the method, the method returns error.
	Find(userID int) (*models.User, error)

	IsUserExist(ctx context.Context, userID int) (bool, error)

	// FindByEmail returns service.ErrUserNotFound if user doesn't exist,
	//returns service.ErrInvalidUserEmail if email is ""
	//or another one if unknown error occurred
	FindByEmail(string) (*models.User, error)

	// FindByLogin returns service.ErrUserNotFound if user doesn't exist,
	//returns service.ErrInvalidUserEmail if login is ""
	//or another one if unknown error occurred
	FindByLogin(login string) (*models.User, error)
}

type TaskService interface {
	CreateGroupTask(ctx context.Context, task *models.Task) error
	CreateUserTask(ctx context.Context, task *models.Task) error

	GetTasksOfGroup(ctx context.Context, groupID, userID int) ([]models.Task, error)
	GetTasksOfUser(ctx context.Context, userID int) ([]models.Task, error)
	GetUserLocalTasks(ctx context.Context, userID int) ([]models.Task, error)

	GetGroupTaskWithContext(ctx context.Context, groupID, taskID int) (*models.Task, error)

	Find(ctx context.Context, taskID int) (*models.Task, error)
	GetAllTasks(ctx context.Context, limit, offset int) ([]models.Task, error)
	GetAllUserTasks(ctx context.Context, userID, limit, offset int) ([]models.Task, error)
}

type UniversityService interface {
	Create(university *models.University) error
	Find(universityID int) (*models.University, error)
}

type GroupService interface {
	Create(group *models.Group, user *models.User) error
	GetAllGroups(limit, offset int) ([]models.Group, error)

	// Find returns *models.Group by groupID or nil if group not found.
	//	If group was not found, method returns service.ErrGroupNotFound.
	//	If an error occurs during the execution of the method, the method returns error.
	Find(groupID int) (*models.Group, error)

	// FindByName returns *models.Group by name or nil if group not found.
	//	If group was not found, method returns service.ErrGroupNotFound.
	//	If an error occurs during the execution of the method, the method returns error.
	FindByName(name string) (*models.Group, error)
	Update(groupID int, ud *models.UpdateGroup) error
	Delete(groupID int, userID int) error

	GetGroupMembers(groupID int) ([]models.User, error)
	GetGroupsUserMemberOf(userID int) ([]models.Group, error)
	IsUserGroupMember(userID, groupID int) (bool, error)
	GetUserPermissions(userID, groupID int) error

	AddUserToGroupByInvite(userID int, invite string) error
	GetInviteLink(groupID int) (*models.GroupInvite, error)
	GetOrCreateInviteLink(groupID int, inviterID int) (*models.GroupInvite, error)
}

type SubjectService interface {
	Create(subject *models.Subject) error
	GetAllSubjects(limit, offset int) ([]models.Subject, error)
	Find(subjectID int) (*models.Subject, error)
	Delete(subjectID int) (*models.Subject, error)
}
