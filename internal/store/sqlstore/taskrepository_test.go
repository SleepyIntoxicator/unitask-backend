package sqlstore

import (
	"backend/internal/api/v1/models"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTaskRepository_CreateGroupTask(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user", "task", "group", "subject", "taskongroup")

	group := models.TestGroup(t)
	subject := models.TestSubject(t)

	s := New(db)
	tester, err := s.User().CreateTester()
	assert.NoError(t, err)

	_ = s.Group().Create(group)
	_ = s.Subject().Create(subject)

	task := models.TestTask(t)
	task.SubjectID = subject.ID
	task.GroupsID = append(task.GroupsID, group.ID)
	task.AddedByID = tester.ID

	assert.NoError(t, s.Task().CreateGroupTask(task))
}

func TestTaskRepository_Find(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user", "task", "group", "subject", "taskongroup", "taskonuser")

	//Creating testing models for task
	task := models.TestTask(t)
	groups := models.TestGroups(t)
	subject := models.TestSubject(t)

	s := New(db)

	//Creating a user who adds this task
	user, err := s.User().CreateTester()
	assert.NoError(t, err)
	assert.NoError(t, s.Subject().Create(subject))

	//Adding groups to the task
	for i := range groups {
		err = s.Group().Create(&groups[i])
		assert.NoError(t, err)

		grForTask, err := s.Group().FindByName(groups[i].CustomName)
		assert.NoError(t, err)

		task.GroupsID = append(task.GroupsID, grForTask.ID)
	}
	task.UsersID = append(task.UsersID, user.ID)

	assert.NoError(t, s.Task().CreateGroupTask(task))
	fmt.Printf("Task is %#v\n", task)

	t2, err := s.Task().Find(task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, t2)
	assert.NotNil(t, t2.GroupsID)
	assert.NotNil(t, t2.UsersID)
	fmt.Printf("Task is %#v\n", t2)
}

func TestTaskRepository_Update(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user", "task", "group", "subject", "taskongroup", "taskonuser")

	s := New(db)

	//Creating testing models for task
	task := models.TestTask(t)
	groups := models.TestGroups(t)
	subject := models.TestSubject(t)

	//Creating a user who adds this task
	user, err := s.User().CreateTester()
	assert.NoError(t, err)
	assert.NoError(t, s.Subject().Create(subject))

	//Adding groups to the task
	for i := range groups {
		err = s.Group().Create(&groups[i])
		assert.NoError(t, err)

		grForTask, err := s.Group().FindByName(groups[i].CustomName)
		assert.NoError(t, err)

		task.GroupsID = append(task.GroupsID, grForTask.ID)
	}
	task.UsersID = append(task.UsersID, user.ID)

	assert.NoError(t, s.Task().CreateGroupTask(task))
	fmt.Printf("Task is %#v\n", task)

}
