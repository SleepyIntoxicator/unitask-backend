package teststore_test

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
	"back-end/internal/app/store/teststore"
	"github.com/stretchr/testify/assert"
	"testing"
	//"back-end/internal/app/store"
)

func TestUserRepository_Create(t *testing.T) {
	s := teststore.New()
	u := models.TestUser(t)
	assert.NoError(t, s.User().Create(u))
	assert.NotNil(t, u)
}

func TestUserRepository_CreateMultipleUsers(t *testing.T) {
	s := teststore.New()
	us := models.TestUsers(t)
	for _, u := range us {
		assert.NoError(t, s.User().Create(&u))
		assert.NotNil(t, u)
	}
}

func TestUserRepository_Find(t *testing.T) {
	s := teststore.New()
	u1 := models.TestUser(t)
	_, err := s.User().Find(u1.ID)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	_ = s.User().Create(u1)
	u2, err := s.User().Find(u1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, u2)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	s := teststore.New()
	u1 := models.TestUser(t)
	_, err := s.User().FindByEmail(u1.Email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	_ = s.User().Create(u1)
	u2, err := s.User().FindByEmail(u1.Email)
	assert.NoError(t, err)
	assert.NotNil(t, u2)

	/*email := "user@example.com"
	_, err := s.User().FindByEmail(email)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

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
