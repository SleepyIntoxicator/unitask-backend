package teststore

import (
	"backend/internal/api/v1/models"
)

type TaskRepository struct {
	store *Store
}

func (r *TaskRepository) CreateGroupTask(task *models.Task) error {
	panic("implement me")
}

func (r *TaskRepository) CreateUserTask(task *models.Task) error {
	panic("implement me")
}

func (r *TaskRepository) GetAll(limit, offset int) ([]models.Task, error) {
	panic("implement me")
}

func (r *TaskRepository) Find(i int) (*models.Task, error) {
	panic("implement me")
}

func (r *TaskRepository) Update(taskID int, updTask *models.UpdateTask) error {
	panic("implement me")
}

func (r *TaskRepository) AssignSubtask(taskID, subtaskID int) error {
	panic("implement me")
}

func (r *TaskRepository) AssignTaskSequence(taskID, nextTaskID int) error {
	panic("implement me")
}

func (r *TaskRepository) AssignTaskToGroup(taskID, groupID int) error {
	panic("implement me")
}

func (r *TaskRepository) AssignTaskToUser(taskID, userID int) error {
	panic("implement me")
}

func (r *TaskRepository) FindPrevTasks(taskID int) ([]int, error) {
	panic("implement me")
}

func (r *TaskRepository) FindNextTasks(taskID int) ([]int, error) {
	panic("implement me")
}

func (r *TaskRepository) FindParentTask(taskID int) (int, error) {
	panic("implement me")
}

func (r *TaskRepository) FindSubtasks(taskID int) ([]int, error) {
	panic("implement me")
}

func (r *TaskRepository) FindTasksOnGroup(groupID int) ([]models.Task, error) {
	panic("implement me")
}

func (r *TaskRepository) FindTasksOnUser(userID int) ([]models.Task, error) {
	panic("implement me")
}

func (r *TaskRepository) FindUserLocalTasks(userID int) ([]models.Task, error) {
	panic("implement me")
}

func (r *TaskRepository) FindGroupsOnTask(taskID int) ([]int, error) {
	panic("implement me")
}

func (r *TaskRepository) FindUsersOnTask(taskID int) ([]int, error) {
	panic("implement me")
}

func (r *TaskRepository) RemoveGroupFromTask(taskID, groupID int) error {
	panic("implement me")
}

func (r *TaskRepository) RemoveUserFromTask(taskID, userID int) error {
	panic("implement me")
}

func (r *TaskRepository) DeleteTask(id int) error {
	panic("implement me")
}
