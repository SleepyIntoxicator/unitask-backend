package store

import (
	"backend/internal/api/v1/models"
	"context"
	"github.com/google/uuid"
)

type AuthRepository interface {
	RegisterApp(ctx context.Context, app *models.RegisteredApp) error
	DeleteApp(ctx context.Context, appUUID uuid.UUID) error
	GetApp(ctx context.Context, appUUID uuid.UUID) (*models.RegisteredApp, error)
	GetAppByName(name string) (*models.RegisteredApp, error)

	AddAppToken(ctx context.Context, t *models.AppToken) error
	RemoveAppTokens(ctx context.Context, appUUID uuid.UUID) error
	GetAppTokenInfo(ctx context.Context, token string) (*models.AppToken, error)
	//Returns register app info
	GetAppTokenByAppUUID(ctx context.Context, appUUID uuid.UUID) (*models.AppToken, error)

	AddUserToken(ctx context.Context, t *models.UserToken) error
	GetUserToken(ctx context.Context, userID int) (*models.UserToken, error)
	RemoveUserTokens(ctx context.Context, userID int) (int, error)
	ClearUserTokens(ctx context.Context) error

	SetUserTokenInvalidByToken(ctx context.Context, token string) error
	SetUserTokenInvalidByUserID(ctx context.Context, userID int) error
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetAll(ctx context.Context, limit, offset int) ([]models.User, error)
	Find(ctx context.Context, userId int) (*models.User, error)
	FindByLogin(ctx context.Context, login string) (*models.User, error)

	// FindByEmail returns store.ErrRecordNotFound if db driver returned sql.ErrNoRows
	//or another error if unknown one occurred.
	FindByEmail(ctx context.Context, email string) (*models.User, error)

	// CreateTester create user for testing
	CreateTester() (*models.User, error)

	IsUserExist(ctx context.Context, userID int) (bool, error)
	GetUserRoles(ctx context.Context, userID int) ([]models.Role, error)
	GetUserToken(ctx context.Context, userID int) (*models.UserToken, error)

	/*FindByFullName(string) (*models.User, error)*/
}

type UniversityRepository interface {
	Create(university *models.University) error
	Find(universityID int) (*models.University, error)
}

type GroupRepository interface {
	Create(*models.Group) error
	GetAllGroups(limit, offset int) ([]models.Group, error)
	Find(int) (*models.Group, error)
	FindByName(string) (*models.Group, error)
	Update(int, *models.UpdateGroup) error
	Delete(int) error

	IsGroupExist(groupID int) (bool, error)

	AddGroupMember(userID, groupID int, inviterID int) error

	IsUserGroupMember(userID, groupID int) (bool, error)
	GetGroupsUserMemberOf(userID int) ([]models.Group, error) //Get IDs of groups that this user is a member of
	GetGroupMembers(groupID int) ([]models.User, error)
	GetMembersCount(groupID int) (int, error)
	GetMemberRoles(userID, groupID int) ([]models.Role, error) //Get the roles that this user have
	GetRolePermissions(roleID int) ([]models.Permission, error)

	GetRole(roleID int) (*models.Role, error)
	GetRoleByName(roleName string) (*models.Role, error)

	GetGroupInvite(groupID int) (*models.GroupInvite, error)
	GetGroupInviteByHash(inviteHash string) (*models.GroupInvite, error)
	AddGroupInviteHash(invite *models.GroupInvite) error
	DeleteGroupInvites(groupID int) error
	DeleteGroupInviteByHash(hash string) error
}

type SubjectRepository interface {
	Create(*models.Subject) error
	GetAll(limit, offset int) ([]models.Subject, error)
	Find(int) (*models.Subject, error)
	Delete(int) (*models.Subject, error)
}

type TaskRepository interface {
	CreateGroupTask(*models.Task) error
	CreateUserTask(*models.Task) error
	GetAll(limit, offset int) ([]models.Task, error)
	Find(int) (*models.Task, error)

	Update(taskID int, updTask *models.UpdateTask) error

	AssignSubtask(taskID, subtaskID int) error
	AssignTaskSequence(taskID, nextTaskID int) error
	AssignTaskToGroup(taskID, groupID int) error
	AssignTaskToUser(taskID, userID int) error

	FindPrevTasks(taskID int) ([]int, error)
	FindNextTasks(taskID int) ([]int, error)
	FindParentTask(taskID int) (int, error)
	FindSubtasks(taskID int) ([]int, error)

	FindTasksOnGroup(groupID int) ([]models.Task, error)
	FindTasksOnUser(userID int) ([]models.Task, error)
	FindUserLocalTasks(userID int) ([]models.Task, error)
	FindGroupsOnTask(taskID int) ([]int, error)
	FindUsersOnTask(taskID int) ([]int, error)

	RemoveGroupFromTask(taskID, groupID int) error
	RemoveUserFromTask(taskID, userID int) error

	DeleteTask(id int) error
	//CreateSubtask(task *models.Task) error
	//AddTaskStatusType(name string) (int, error)
	//AddTaskStatus()
}

type TaskStatusTypeRepository interface {
	Create(statusType *models.TaskStatusType) error
	GetAllTypes() (*[]models.TaskStatusType, error)
}

type TaskStatusRepository interface {
	Create(taskStatus *models.TaskStatus) error
	Get(taskStatusID int) (*models.TaskStatus, error)
}

type LocalTaskRepository interface {
	Create(userTask *models.UserTask) error
	GetLocalTasks(userID int) ([]models.UserTask, error)
	GetLocalTask(userID int, taskID int) (*models.UserTask, error)
	//ChangeTask()
	ChangeTaskStatus(taskID int, taskStatus *models.TaskStatus) error
	DeleteTask(userID int, taskID int) error
}
