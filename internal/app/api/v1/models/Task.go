package models

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"time"
)

/*
	Задачи.
		Разделяются на:
			Групповые
				Выданы на группу и клонируются у участников.
			Пользовательские
				Выданы на пользователя\введены пользователем.
			Одновременно может быть лишь одно состояние.
		Для клонирования отдельная структура UserTask
		Может иметь ссылку на родительскую задачу ParentTaskID.
		Может иметь ссылку на следующие задачу NextTasksIDs.
*/
type Task struct {
	ID          int  `json:"id"`
	TypeID      int  `json:"type_id,omitempty" db:"type_id"`
	IsGroupTask bool `json:"is_group_task" db:"is_task_group"`
	IsLocalTask bool `json:"is_local_task" db:"is_task_local"`

	Name    string    `json:"name" db:"name"`
	Content string    `json:"content" db:"content"`
	StartAt time.Time `json:"start_at" db:"start_at"`
	EndAt   time.Time `json:"end_at" db:"end_at"`

	GroupsID     []int `json:"groups_ids"`
	UsersID      []int `json:"users_ids"`
	SubjectID    int   `json:"subject_id" db:"subject_id"`
	ParentTaskID int   `json:"parent_task_id" db:"parent_task_id"`
	SubtasksIDs  []int `json:"subtasks_ids"`
	PrevTasksIDs []int `json:"prev_tasks_ids" db:"prev_task_id"`
	NextTasksIDs []int `json:"next_tasks_ids" db:"next_task_id"`
	AddedByID    int   `json:"added_by_id" db:"added_by_id"`

	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at" db:"updated_at"`
	UpdatesCount  int       `json:"updates_count" db:"updates_count"`
	Views         int       `json:"watches" db:"views"`
}

func (t *Task) Validate() error {
	if t.IsGroupTask && t.IsLocalTask {
		return errors.New("the task can't be in two states. ")
	}
	for _, id := range t.PrevTasksIDs {
		if id < 1 {
			return errors.New("previous task can't have zero id or less")
		}
		if id == t.ID {
			return errors.New("the task can't point to itself")
		}
	}

	for _, id := range t.NextTasksIDs {
		if id < 1 {
			return errors.New("next task can't have zero id or less")
		}
		if id == t.ID {
			return errors.New("the task can't point to itself")
		}
	}

	if t.ParentTaskID < 0 {
		return errors.New("the ID of the parent task can't be less than zero")
	}
	if t.ParentTaskID != 0 && t.ParentTaskID == t.ID {
		return errors.New("the task can't point to itself")
	}
	err := validation.ValidateStruct(
		t,
		validation.Field(&t.SubjectID, validation.Required, validation.NotNil),
		validation.Field(&t.AddedByID, validation.Required, validation.NotNil),
		validation.Field(&t.Name, validation.Required, validation.RuneLength(1, 64)),
		validation.Field(&t.Content, validation.Required, validation.NotNil))
	//validation.Field(&t.GroupID, validation.Required, validation.Min(0)),
	//validation.Field(&t.UserID, validation.Required, validation.Min(0)),
	//validation.Field(&t.ParentTaskID, validation.Required, validation.NotNil),
	//validation.Field(&t.NextTasksIDs, validation.Required, validation.NotNil),

	return err
}

type UpdateTask struct {
	Name      *string
	Content   *string
	StartAt   *time.Time
	EndAt     *time.Time
	SubjectID *int
}

// TaskParams...
type TaskParams struct {
	ExpectSubmittingReport bool `json:"expect_submitting_report"`
	ExpectVerification     bool `json:"expect_verification"`
	ExpectRevision         bool `json:"expect_revision"`
}

/*
	Ex:
		0000
		00000000
		true
		0001
		null

		0001
		00000000
		false
		0002

*/

type UserTask struct {
	ID           int
	UserID       int  `json:"user_id"`
	IsLocal      bool `json:"is_local"`
	TaskStatusID int  `json:"task_status_id"`
	ParentTaskID int  `json:"parent_task_id"`
}

/*
TaskStatus - is the struct of represent
	Ex:
		0000
		1_task
		task_in_process
		"Task in progress ..."

		0001
		1_task
		task_done
		"The task is done"

		0002
		2_task
		task_expired
		"The issue has expired."
*/
type TaskStatus struct {
	ID               int
	TaskStatusTypeID int    `json:"task_status_type_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
}

/*
TaskStatusType - is the struct of
	Ex:
		__		- base (for all tasks)
		task	- general
		rep		- if report	 expected
		protect	- if protect expected
*/
type TaskStatusType struct {
	ID   int
	Name string `json:"name"`
}
