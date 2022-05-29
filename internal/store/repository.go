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
	Create(ctx context.Context, university *models.University) error
	Find(ctx context.Context, universityID int) (*models.University, error)
}

type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) error
	GetAllGroups(ctx context.Context, limit, offset int) ([]models.Group, error)
	Find(ctx context.Context, id int) (*models.Group, error)
	FindByName(ctx context.Context, name string) (*models.Group, error)
	Update(ctx context.Context, groupID int, up *models.UpdateGroup) error
	Delete(ctx context.Context, id int) error

	IsGroupExist(ctx context.Context, groupID int) (bool, error)

	AddGroupMember(ctx context.Context, userID, groupID int, inviterID int) error

	IsUserGroupMember(ctx context.Context, userID, groupID int) (bool, error)
	GetGroupsUserMemberOf(ctx context.Context, userID int) ([]models.Group, error) //Get IDs of groups that this user is a member of
	GetGroupMembers(ctx context.Context, groupID int) ([]models.User, error)
	GetMembersCount(ctx context.Context, groupID int) (int, error)
	GetMemberRoles(ctx context.Context, userID, groupID int) ([]models.Role, error) //Get the roles that this user have
	GetRolePermissions(ctx context.Context, roleID int) ([]models.Permission, error)

	GetRole(ctx context.Context, roleID int) (*models.Role, error)
	GetRoleByName(ctx context.Context, roleName string) (*models.Role, error)

	GetGroupInvite(ctx context.Context, groupID int) (*models.GroupInvite, error)
	GetGroupInviteByHash(ctx context.Context, inviteHash string) (*models.GroupInvite, error)
	AddGroupInviteHash(ctx context.Context, invite *models.GroupInvite) error
	DeleteGroupInvites(ctx context.Context, groupID int) error
	DeleteGroupInviteByHash(ctx context.Context, hash string) error
}

type SubjectRepository interface {
	Create(ctx context.Context, subject *models.Subject) error
	GetAll(ctx context.Context, limit, offset int) ([]models.Subject, error)
	Find(ctx context.Context, id int) (*models.Subject, error)
	Delete(ctx context.Context, id int) (*models.Subject, error)
}

type TaskRepository interface {
	CreateGroupTask(ctx context.Context, task *models.Task) error
	CreateUserTask(ctx context.Context, task *models.Task) error
	GetAll(ctx context.Context, limit, offset int) ([]models.Task, error)
	Find(ctx context.Context, id int) (*models.Task, error)

	Update(ctx context.Context, taskID int, updTask *models.UpdateTask) error
	DeleteTask(ctx context.Context, id int) error

	AssignSubtask(ctx context.Context, taskID, subtaskID int) error
	AssignTaskSequence(ctx context.Context, taskID, nextTaskID int) error
	AssignTaskToGroup(ctx context.Context, taskID, groupID int) error
	AssignTaskToUser(ctx context.Context, taskID, userID int) error

	FindNextTasks(ctx context.Context, taskID int) ([]int, error)
	FindPrevTasks(ctx context.Context, taskID int) ([]int, error)
	FindSubtasks(ctx context.Context, taskID int) ([]int, error)
	FindParentTask(ctx context.Context, taskID int) (int, error)

	FindTasksOnGroup(ctx context.Context, groupID int) ([]models.Task, error)
	FindTasksOnUser(ctx context.Context, userID int) ([]models.Task, error)
	FindUserLocalTasks(ctx context.Context, userID int) ([]models.Task, error)
	FindGroupsOnTask(ctx context.Context, taskID int) ([]int, error)
	FindUsersOnTask(ctx context.Context, taskID int) ([]int, error)

	RemoveGroupFromTask(ctx context.Context, taskID, groupID int) error
	RemoveUserFromTask(ctx context.Context, taskID, userID int) error

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
