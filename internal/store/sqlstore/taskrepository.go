package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type TaskRepository struct {
	store *Store
}

func (r *TaskRepository) CreateGroupTask(t *models.Task) error {
	now := time.Now()

	query := `INSERT INTO task (type_id, is_task_group, is_task_local, name, content, start_at, end_at, subject_id, added_by_id, created_at, updated_at, updates_count, views) 
				VALUES ($1, true, false, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, is_task_group, is_task_local, created_at, updated_at`
	if err := r.store.db.QueryRow(
		query,
		t.TypeID,
		t.Name,
		t.Content,
		t.StartAt,
		t.EndAt,
		t.SubjectID,
		t.AddedByID,
		now,
		now,
		0,
		0,
	).Scan(
		&t.ID,
		&t.IsGroupTask,
		&t.IsLocalTask,
		&t.CreatedAt,
		&t.LastUpdatedAt,
	); err != nil {
		return err
	}

	for _, groupID := range t.GroupsID {
		if err := r.AssignTaskToGroup(t.ID, groupID); err != nil {
			return err
		}
	}

	for _, userID := range t.UsersID {
		if err := r.AssignTaskToUser(t.ID, userID); err != nil {
			return err
		}
	}

	if t.ParentTaskID != 0 {
		if err := r.AssignSubtask(t.ID, t.ParentTaskID); err != nil {
			return err
		}
	}

	for _, taskID := range t.PrevTasksIDs {
		if err := r.AssignTaskSequence(taskID, t.ID); err != nil {
			return err
		}
	}

	for _, taskID := range t.NextTasksIDs {
		if err := r.AssignTaskSequence(t.ID, taskID); err != nil {
			return err
		}
	}

	return nil
}

func (r *TaskRepository) CreateGroupTaskOnTestRequireSpeedTest(t *models.Task) error {
	if err := t.Validate(); err != nil {
		return err
	}

	now := time.Now()

	tx := r.store.db.MustBegin()
	query := `INSERT INTO task (type_id, is_task_group, is_task_local, name, content, start_at, end_at, subject_id, added_by_id, created_at, updated_at, updates_count, views) 
						values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`
	if err := tx.QueryRow(
		query,
		t.TypeID,
		true,
		false,
		t.Name,
		t.Content,
		t.StartAt,
		t.EndAt,
		t.SubjectID,
		t.AddedByID,
		now,
		now,
		0,
		0,
	).Scan(&t.ID); err != nil {
		return err
	}
	// range := AssignTaskToGroup
	for _, groupID := range t.GroupsID {
		if _, err := tx.Exec(
			"INSERT INTO taskongroup (task_id, group_id) VALUES ($1, $2)",
			t.ID,
			groupID,
		); err != nil {
			return err
		}
	}

	for _, groupID := range t.GroupsID {
		err := r.assignTaskToGroupWithTx(tx, t.ID, groupID)
		if err != nil {
			return err
		}
	}

	// range := AssignTaskToUser
	for _, userID := range t.UsersID {
		if _, err := tx.Exec(
			"INSERT INTO taskonuser (task_id, user_id) VALUES ($1, $2)",
			t.ID,
			userID,
		); err != nil {
			return err
		}
	}

	// AssignSubtask( t.ID, t.ParentTaskID )
	if t.ParentTaskID != 0 {
		if _, err := r.store.db.Exec(
			"INSERT INTO subtask (task_id, parent_task_id) VALUES ($1, $2)",
			t.ID,
			t.ParentTaskID,
		); err != nil {
			return err
		}
	}
	panic("t.PrevTaskID\\t.NextTaskID was changed to slice and this code doesn't work")
	// AssignTaskSequence( t.ID, t.NextTasksIDs )
	if _, err := tx.Exec(
		"INSERT INTO tasktree (task_id, next_task_id) VALUES ($1, $2)",
		t.ID,
		t.NextTasksIDs,
	); err != nil {
		return err
	}

	for _, taskID := range t.PrevTasksIDs {
		if err := r.AssignTaskSequence(taskID, t.ID); err != nil {
			return err
		}
	}

	for _, taskID := range t.NextTasksIDs {
		if err := r.AssignTaskSequence(t.ID, taskID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *TaskRepository) CreateUserTask(t *models.Task) error {
	if err := t.Validate(); err != nil {
		return err
	}

	now := time.Now()

	query := `INSERT INTO task (type_id, is_task_group, is_task_local, name, content, start_at, end_at, subject_id, added_by_id, created_at, updated_at, updates_count, views) 
						values ($1, false, true, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, is_task_group, is_task_local, created_at, updated_at`
	if err := r.store.db.QueryRow(
		query,
		t.TypeID,
		t.Name,
		t.Content,
		t.StartAt,
		t.EndAt,
		t.SubjectID,
		t.AddedByID,
		now,
		now,
		0,
		0,
	).Scan(
		&t.ID,
		&t.IsGroupTask,
		&t.IsLocalTask,
		&t.CreatedAt,
		&t.LastUpdatedAt); err != nil {
		return err
	}

	for _, userID := range t.UsersID {
		if err := r.AssignTaskToUser(t.ID, userID); err != nil {
			return err
		}
	}

	if t.ParentTaskID != 0 {
		if err := r.AssignSubtask(t.ID, t.ParentTaskID); err != nil {
			return err
		}
	}

	for _, taskID := range t.PrevTasksIDs {
		if err := r.AssignTaskSequence(taskID, t.ID); err != nil {
			return err
		}
	}

	for _, taskID := range t.NextTasksIDs {
		if err := r.AssignTaskSequence(t.ID, taskID); err != nil {
			return err
		}
	}

	return nil
}

func (r *TaskRepository) AssignSubtask(taskID int, parentTaskID int) error {
	if taskID == parentTaskID {
		return models.ErrTaskCannotPointToItself
	}
	_, err := r.store.db.Exec(
		"INSERT INTO subtask (task_id, parent_task_id) VALUES ($1, $2)",
		taskID,
		parentTaskID,
	)
	return err
}

func (r *TaskRepository) AssignTaskSequence(taskID int, nextTaskID int) error {
	if taskID == nextTaskID {
		return models.ErrTaskCannotPointToItself
	}
	_, err := r.store.db.Exec(
		"INSERT INTO tasktree (task_id, next_task_id) VALUES ($1, $2)",
		taskID,
		nextTaskID,
	)

	return err
}

func (r *TaskRepository) AssignTaskToGroup(taskID int, groupID int) error {
	_, err := r.store.db.Exec(
		"INSERT INTO taskongroup (task_id, group_id) VALUES ($1, $2)",
		taskID,
		groupID,
	)
	return err
}

func (r *TaskRepository) assignTaskToGroupWithTx(tx *sqlx.Tx, taskID, groupID int) error {
	_, err := tx.Exec(
		"INSERT INTO taskongroup (task_id, group_id) VALUES ($1, $2)",
		taskID,
		groupID,
	)
	return err
}

func (r *TaskRepository) AssignTaskToUser(taskID int, userID int) error {
	_, err := r.store.db.Exec(
		"INSERT INTO taskonuser (task_id, user_id) VALUES ($1, $2)",
		taskID,
		userID,
	)
	return err
}

func (r *TaskRepository) assignTaskToUserWithTx(tx *sqlx.Tx, taskID, userID int) error {
	_, err := tx.Exec(
		"INSERT INTO taskonuser (task_id, user_id) VALUES ($1, $2)",
		taskID,
		userID,
	)
	return err
}

func (r *TaskRepository) GetAll(limit, offset int) ([]models.Task, error) {
	var tasks []models.Task

	query := `SELECT id, type_id, is_task_group, is_task_local, name, content, start_at, end_at,
					subject_id, added_by_id, created_at, updated_at, updates_count, views 
				FROM task ORDER BY id`
	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	rows, err := r.store.db.Query(query)
	if err != nil {
		//TODO: create and delete PgError check
		return nil, store.HandleErrorNoRows(err)
	}

	for rows.Next() {
		task := models.Task{}
		if err := rows.Scan(
			&task.ID,
			&task.TypeID,
			&task.IsGroupTask,
			&task.IsLocalTask,

			&task.Name,
			&task.Content,
			&task.StartAt,
			&task.EndAt,
			&task.SubjectID,

			&task.AddedByID,

			&task.CreatedAt,
			&task.LastUpdatedAt,
			&task.UpdatesCount,
			&task.Views,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	rows.Close()

	for i, task := range tasks {
		tasks[i].GroupsID, err = r.store.Task().FindGroupsOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].UsersID, err = r.store.Task().FindUsersOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].ParentTaskID, err = r.store.Task().FindParentTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].SubtasksIDs, err = r.store.Task().FindSubtasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].PrevTasksIDs, err = r.store.Task().FindPrevTasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}
	}

	return tasks, nil
}

func (r *TaskRepository) FindTasksOnGroup(groupId int) ([]models.Task, error) {
	var tasks []models.Task
	query := `SELECT * FROM task WHERE is_task_group = true AND id IN 
                            (SELECT task_id FROM taskongroup WHERE group_id = $1) ORDER BY id`

	err := r.store.db.Select(&tasks, query, groupId)

	for i, task := range tasks {
		tasks[i].GroupsID, err = r.store.Task().FindGroupsOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].UsersID, err = r.store.Task().FindUsersOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].ParentTaskID, err = r.store.Task().FindParentTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].SubtasksIDs, err = r.store.Task().FindSubtasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].PrevTasksIDs, err = r.store.Task().FindPrevTasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}
	}

	return tasks, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) FindTasksOnUser(userID int) ([]models.Task, error) {
	var tasks []models.Task
	query := `SELECT * FROM task WHERE id IN 
                            (SELECT task_id FROM taskonuser WHERE user_id = $1 ) ORDER BY id`

	err := r.store.db.Select(&tasks, query, userID)

	for i, task := range tasks {
		tasks[i].GroupsID, err = r.store.Task().FindGroupsOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].UsersID, err = r.store.Task().FindUsersOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].ParentTaskID, err = r.store.Task().FindParentTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}

	}

	return tasks, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) FindUserLocalTasks(userID int) ([]models.Task, error) {
	var tasks []models.Task
	query := `SELECT * FROM task WHERE is_task_local = true AND id IN 
                            (SELECT task_id FROM taskonuser WHERE user_id = $1 ) ORDER BY id`

	err := r.store.db.Select(&tasks, query, userID)

	for i, task := range tasks {
		tasks[i].UsersID, err = r.store.Task().FindUsersOnTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].ParentTaskID, err = r.store.Task().FindParentTask(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}

		tasks[i].NextTasksIDs, err = r.store.Task().FindNextTasks(task.ID)
		if err != nil {
			return nil, err
		}

	}

	return tasks, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) FindGroupsOnTask(taskID int) ([]int, error) {
	var groupsIDs []int
	query := `SELECT group_id FROM taskongroup WHERE task_id = $1 ORDER BY group_id`

	err := r.store.db.Select(&groupsIDs, query, taskID)
	return groupsIDs, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) FindUsersOnTask(taskID int) ([]int, error) {
	var usersIDs []int
	query := `SELECT user_id FROM taskonuser WHERE task_id = $1 ORDER BY user_id`
	err := r.store.db.Select(&usersIDs, query, taskID)
	return usersIDs, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) Find(id int) (*models.Task, error) {
	t := &models.Task{}

	query := `SELECT id, type_id, is_task_group, is_task_local, name, content, start_at, end_at, subject_id, added_by_id, created_at, updated_at, updates_count, views 
				FROM task WHERE id = $1`
	err := r.store.db.QueryRow(query, id).Scan(
		&t.ID,
		&t.TypeID,
		&t.IsGroupTask,
		&t.IsLocalTask,

		&t.Name,
		&t.Content,
		&t.StartAt,
		&t.EndAt,
		&t.SubjectID,

		&t.AddedByID,

		&t.CreatedAt,
		&t.LastUpdatedAt,
		&t.UpdatesCount,
		&t.Views,
	)
	if err != nil {
		return nil, store.HandleErrorNoRows(err)
	}

	t.GroupsID, err = r.FindGroupsOnTask(t.ID)
	if err != nil {
		return nil, err
	}

	t.UsersID, err = r.FindUsersOnTask(t.ID)
	if err != nil {
		return nil, err
	}

	// Finding parent task. If the row isn't found, the task has no parent and parent_task_id = 0
	t.ParentTaskID, err = r.FindParentTask(t.ID)
	if err != nil {
		return nil, err
	}

	t.SubtasksIDs, err = r.FindSubtasks(t.ID)
	if err != nil {
		return nil, err
	}

	// Finding next task. If the row isn't found, the task has no next task and next_task_id = 0
	t.NextTasksIDs, err = r.FindNextTasks(t.ID)
	if err != nil {
		return nil, err
	}

	t.PrevTasksIDs, err = r.FindPrevTasks(t.ID)
	if err != nil {
		return nil, err
	}

	query = `UPDAtE task SET views = views + 1 WHERE id = $1 RETURNING views`
	err = r.store.db.QueryRow(query, t.ID).Scan(&t.Views)
	if err != nil {
		return nil, err
	}

	return t, err
}

func (r *TaskRepository) Update(taskID int, updTask *models.UpdateTask) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argID := 1

	if updTask.Name != nil {
		setValues = append(setValues, "name=$%d")
		args = append(args, updTask.Name)
		argID++
	}
	if updTask.Content != nil {
		setValues = append(setValues, "content=$%d")
		args = append(args, updTask.Content)
		argID++
	}
	if updTask.StartAt != nil {
		setValues = append(setValues, "start_at=$%d")
		args = append(args, updTask.StartAt)
		argID++
	}
	if updTask.EndAt != nil {
		setValues = append(setValues, "end_at=$%d")
		args = append(args, updTask.EndAt)
		argID++
	}
	if updTask.SubjectID != nil {
		//Checking for the existence of an item in the db
		var i int
		if err := r.store.db.Select(&i, "SELECT id FROM task WHERE id = $1", updTask.SubjectID); err != nil {
			return err
		}

		setValues = append(setValues, "subject_id=$%d")
		args = append(args, updTask.SubjectID)
		argID++
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf("UPDATE task SET %s WHERE id = $%d", setQuery, argID)
	args = append(args, taskID)

	_, err := r.store.db.Exec(query, args...)
	return err
}

func (r *TaskRepository) FindParentTask(id int) (int, error) {
	parentTaskID := 0

	err := r.store.db.QueryRow(
		"SELECT parent_task_id FROM subtask WHERE task_id = $1", id,
	).Scan(&parentTaskID)

	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return parentTaskID, nil
}

func (r *TaskRepository) FindSubtasks(taskID int) ([]int, error) {
	var subtasksIDs []int

	err := r.store.db.Select(&subtasksIDs,
		"SELECT task_id FROM subtask WHERE parent_task_id = $1", taskID,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return subtasksIDs, nil
}

func (r *TaskRepository) FindNextTasks(id int) ([]int, error) {
	var nextTasksIDs []int
	err := r.store.db.Select(&nextTasksIDs, `SELECT next_task_id FROM tasktree WHERE task_id = $1`, id)
	return nextTasksIDs, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) FindPrevTasks(taskID int) ([]int, error) {
	var prevTasksIDs []int
	err := r.store.db.Select(&prevTasksIDs, `SELECT task_id FROM tasktree WHERE next_task_id = $1`,
		taskID)
	return prevTasksIDs, store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) RemoveGroupFromTask(taskID, groupID int) error {
	_, err := r.store.db.Exec("DELETE FROM taskongroup WHERE task_id = $1 AND group_id = $2", taskID, groupID)
	return err
}

func (r *TaskRepository) RemoveUserFromTask(taskID, userID int) error {
	_, err := r.store.db.Exec("DELETE FROM taskonuser WHERE task_id = $1 AND user_id = $2", taskID, userID)
	return err
}

func (r *TaskRepository) DeleteTask(id int) error {
	_, err := r.store.db.Exec("DELETE FROM task WHERE id = $1", id)
	return store.HandleIgnoreErrorNoRows(err)
}

func (r *TaskRepository) AddTaskStatusType(statusType string) error {
	return nil
}
