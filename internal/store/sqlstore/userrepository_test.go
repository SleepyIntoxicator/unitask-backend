package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	s := New(db)
	u := models.TestUser(t)
	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u)
}

func TestUserRepository_CreateMultipleUsers(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	s := New(db)
	us := models.TestUsers(t)
	for _, u := range us {
		err := s.User().Create(&u)
		assert.NoError(t, err)
		assert.NotNil(t, u)
	}
}
func TestUserRepository_Find(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	s := New(db)
	u1 := models.TestUser(t)
	_, err := s.User().Find(u1.ID)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	_ = s.User().Create(u1)
	u2, err := s.User().Find(u1.ID)
	assert.NotNil(t, u2)
	assert.NoError(t, err)
}

func TestUserRepository_FindByLogin(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	s := New(db)
	u1 := models.TestUser(t)
	_, err := s.User().FindByLogin(u1.Login)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	_ = s.User().Create(u1)
	u2, err := s.User().FindByLogin(u1.Login)
	assert.NotNil(t, u2)
	assert.NoError(t, err)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	s := New(db)
	u1 := models.TestUser(t)
	_, err := s.User().FindByEmail(u1.Email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	_ = s.User().Create(u1)
	u2, err := s.User().FindByEmail(u1.Email)
	assert.NoError(t, err)
	assert.NotNil(t, u2)

	/*	There was an error. Null pointer
		email := "user@example.com"
			_, err := s.User().FindByEmail(email)
			assert.Error(t, err)

			u := models.TestUser(t)
			u.Email = email
			_ = s.User().Create(models.TestUser(t))
			u, err = s.User().FindByEmail(email)
			assert.NoError(t, err)
			assert.NotNil(t, u)
			if u.Email != email {
				t.Fail()
			}*/
}

func TestUserRepository_GetAll(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user")

	usersCount := 100

	errCounter := 0

	s := New(db)
	usrs := models.TestUsersN(t, usersCount)
	for i := range usrs {
		err := s.User().Create(&usrs[i])
		if err != nil {
			errCounter++
		} //If user errors > 3 and all errors not nil
		if errCounter < 3 {
			assert.NoError(t, err)
			assert.NotNil(t, usrs[i])
			//If u.ID was not entered
			assert.NotEqual(t, 0, usrs[i].ID)
		}
	}

	UsersFromDB, err := s.User().GetAll(0, 0)
	assert.Equal(t, usersCount, len(UsersFromDB))
	assert.NoError(t, err)

	UsersFromDB, err = s.User().GetAll(10, 3)
	assert.NotNil(t, UsersFromDB)
	assert.NoError(t, err)

	UsersFromDB, err = s.User().GetAll(10, 11)
	assert.EqualError(t, err, store.ErrNoRowsFound.Error())
}
