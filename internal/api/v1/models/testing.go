package models

import (
	"fmt"
	"testing"
	"time"
)

func TestUserForRegister(t *testing.T) *User {
	return &User{
		Login:    "UserExample",
		FullName: "User Example",
		Email:    "user@example.org",
		Password: "password",
	}
}

func TestUser(t *testing.T) *User {
	return &User{
		Login:    "UserExample",
		FullName: "User Example",
		Email:    "user@example.org",
		Password: "password",
	}
}

func TestUsers(t *testing.T) []User {
	testUsers := []User{
		{
			Login:    "UserExample1",
			FullName: "User Example1",
			Email:    "user1@example.org",
			Password: "password1",
		},
		{
			Login:    "UserExample2",
			FullName: "User Example2",
			Email:    "user2@example.org",
			Password: "password2",
		},
		{
			Login:    "UserExample3",
			FullName: "User Example3",
			Email:    "user3@example.org",
			Password: "password3",
		},
		{
			Login:    "UserExample4",
			FullName: "User Example4",
			Email:    "user4@example.org",
			Password: "password4",
		},
	}
	return testUsers
}

func TestUsersN(t *testing.T, usersNumber int) []User {
	var testUsers []User
	for i := 1; i < usersNumber + 1; i++ {
		u := User{
			Login:    fmt.Sprintf("UserExample%d", i),
			FullName: fmt.Sprintf("User Example%d", i),
			Email:    fmt.Sprintf("user%d@example.org", i),
			Password: fmt.Sprintf("password%d", i),
		}
		testUsers = append(testUsers, u)
	}
	return testUsers
}

func TestTask(t *testing.T) *Task {
	return &Task{
		TypeID:       0,
		Name:         "Тестовая задача №1",
		Content:      "Выполнить задачу N1",
		IsGroupTask:   true,
		IsLocalTask:   false,
		GroupsID:      []int{},
		UsersID:       []int{},
		SubjectID:     1,
		ParentTaskID:  0,
		PrevTasksIDs:  nil,
		NextTasksIDs:  nil,
		AddedByID:     1,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
		UpdatesCount:  0,
		Views:         0,
	}
}

func TestTasks(t *testing.T) []Task {
	testTasks := []Task{
		{
			TypeID:      0,
			IsGroupTask: true,
			IsLocalTask: false,

			Name:    "Задача №1",
			Content: "Текст задачи №1",

			GroupsID:     []int{},
			UsersID:      []int{},
			SubjectID:    0,
			ParentTaskID: 0,
			PrevTasksIDs: nil,
			NextTasksIDs: nil,
			AddedByID:    0,
		},
		{
			TypeID:      0,
			IsGroupTask: true,
			IsLocalTask: false,

			Name:    "Задача №2",
			Content: "Текст задачи №2",

			GroupsID:     []int{},
			UsersID:      []int{},
			SubjectID:    0,
			ParentTaskID: 0,
			PrevTasksIDs: nil,
			NextTasksIDs: nil,
			AddedByID:    0,
		},
		{
			TypeID:      0,
			IsGroupTask: true,
			IsLocalTask: false,

			Name:    "Задача №3",
			Content: "Текст задачи №3",

			GroupsID:     []int{},
			UsersID:      []int{},
			SubjectID:    0,
			ParentTaskID: 0,
			PrevTasksIDs: nil,
			NextTasksIDs: nil,
			AddedByID:    0,
		},
	}
	return testTasks
}

func TestGroup(t *testing.T) *Group {
	return &Group{
		CustomName:   "Testing group name",
		UniversityID: 0,
	}
}

func TestGroups(t *testing.T) []Group {
	testGroups := []Group{
		{
			CustomName:   "Группа 1",
			UniversityID: 1,
		},
		{
			CustomName:   "Группа 2",
			UniversityID: 1,
		},
		{
			CustomName:   "Группа 3",
			UniversityID: 1,
		},
		{
			CustomName:   "Группа 4",
			UniversityID: 0,
		},
	}

	return testGroups
}

func TestSubject(t *testing.T) *Subject {
	return &Subject{
		Name: "Testing subject name",
	}
}
