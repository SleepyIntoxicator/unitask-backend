package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
	"sort"
)

type TaskService struct {
	service *Service
}

func (s *TaskService) CreateGroupTask(ctx context.Context, task *models.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	err := s.service.store.Task().CreateGroupTask(task)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) CreateUserTask(ctx context.Context, task *models.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	err := s.service.store.Task().CreateUserTask(task)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) Find(ctx context.Context, taskID int) (*models.Task, error) {
	task, err := s.service.store.Task().Find(taskID)
	if err == store.ErrRecordNotFound {
		return nil, service.ErrTaskNotFound
	}

	return task, err
}

// GetTasksOfGroup Returns: slice of tasks, or error if any
//	Requires: The user must be a member of the group
func (s *TaskService) GetTasksOfGroup(ctx context.Context, groupID, userID int) ([]models.Task, error) {
	// Verifying that the user exists
	isExist, err := s.service.store.User().IsUserExist(userID)
	if err != nil || !isExist {
		return nil, err
	}

	// Verifying that the group exists
	group, err := s.service.Group().Find(groupID)
	if err != nil {
		return nil, err
	}

	isUserMemberOf, err := s.service.Group().IsUserGroupMember(userID, group.ID)
	if err != nil {
		return nil, err
	}

	if !isUserMemberOf {
		return nil, service.ErrUserIsNotGroupMember
	}

	tasks, err := s.service.store.Task().FindTasksOnGroup(groupID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) GetTasksOfUser(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := s.service.store.Task().FindTasksOnUser(userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) GetUserLocalTasks(ctx context.Context, userID int) ([]models.Task, error) {
	tasks, err := s.service.store.Task().FindUserLocalTasks(userID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskService) GetGroupTaskWithContext(ctx context.Context, groupID, taskID int) (*models.Task, error) {
	user, err := s.service.getUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isUserMember, err := s.service.Group().IsUserGroupMember(user.ID, groupID)
	if err != nil {
		return nil, err
	}
	if !isUserMember {
		return nil, models.ErrUserIsNotGroupMember
	}

	tasks, err := s.service.Task().GetTasksOfGroup(ctx, groupID, user.ID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == taskID {
			return &task, nil
		}
	}

	return nil, service.ErrTaskNotFound
}

func (s *TaskService) GetAllTasks(ctx context.Context, limit, offset int) ([]models.Task, error) {
	if limit < 0 || offset < 0 {
		return nil, service.ErrInvalidLimitOrPage
	}

	return s.service.store.Task().GetAll(limit, offset)
}

func (s *TaskService) GetAllUserTasks(ctx context.Context, userID, limit, offset int) ([]models.Task, error) {
	var TasksAvailableToUser []models.Task

	//Adding tasks assigned to groups that the user is member of
	groupsUserMemberOf, err := s.service.Group().GetGroupsUserMemberOf(userID)
	if err != nil {
		return nil, err
	}
	for _, group := range groupsUserMemberOf {
		tasksOfGroup, err := s.service.store.Task().FindTasksOnGroup(group.ID)
		if err != nil {
			return nil, err
		}

		TasksAvailableToUser = append(TasksAvailableToUser, tasksOfGroup...)
	}

	//Adding tasks assigned to current user
	tasksAssignedToUser, err := s.service.store.Task().FindTasksOnUser(userID)
	if err != nil {
		return nil, err
	}
	TasksAvailableToUser = append(TasksAvailableToUser, tasksAssignedToUser...)

	sort.Slice(TasksAvailableToUser, func(i, j int) bool {
		return TasksAvailableToUser[i].ID < TasksAvailableToUser[j].ID
	})

	start := 0
	end := len(TasksAvailableToUser)

	if offset < len(TasksAvailableToUser) {
		start = offset
	} else {
		return TasksAvailableToUser[:0], nil
	}
	if limit != 0 {
		if limit+offset < len(TasksAvailableToUser) {
			end = limit + offset
		} else if offset >= len(TasksAvailableToUser) {
			return TasksAvailableToUser[:0], nil
		}
	}
	return TasksAvailableToUser[start:end], nil
}
