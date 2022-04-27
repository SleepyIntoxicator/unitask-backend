package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupRepository_Create(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("group")

	g := models.TestGroup(t)
	s := New(db)
	assert.NoError(t, s.Group().Create(g))
	assert.NotNil(t, g)
}

func TestGroupRepository_Find(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("group")

	g1 := models.TestGroup(t)
	s := New(db)
	_, err := s.Group().Find(g1.ID)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	assert.NoError(t, s.Group().Create(g1))
	g2, err := s.Group().Find(g1.ID)
	assert.NoError(t, err)
	assert.NotNil(t, g2)
}

func TestGroupRepository_Delete(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("group")

	g := models.TestGroup(t)
	s := New(db)
	assert.NoError(t, s.Group().Create(g))

	uid := 3
	name := "Updated group"
	ug := models.UpdateGroup{
		UniversityID: &uid,
		Name:         &name,
	}
	assert.NoError(t, s.Group().Update(1, &ug))
	g2, err := s.Group().Find(g.ID)
	assert.NoError(t, err)
	assert.NotNil(t, g2)

	assert.NoError(t, s.Group().Delete(g.ID))
}

func TestGroupRepository_Update(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("group")

	g := models.TestGroup(t)
	s := New(db)
	assert.NoError(t, s.Group().Create(g))

	uid := 3
	name := "Updated group"
	ug := models.UpdateGroup{
		UniversityID: &uid,
		Name:         &name,
	}
	assert.NoError(t, s.Group().Update(1, &ug))
	updGroup, err := s.Group().Find(g.ID)
	assert.NoError(t, err)
	assert.Equal(t, g.ID, updGroup.ID)
	assert.Equal(t, *ug.UniversityID, updGroup.UniversityID)
	assert.Equal(t, *ug.Name, updGroup.CustomName)
}

func TestGroupRepository_GetGroupMembers(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("user", "group", "groupmember")

	s := New(db)
	testUsers := models.TestUsers(t)
	for i := range testUsers {
		err := s.User().Create(&testUsers[i])
		assert.NoError(t, err)
		assert.NotNil(t, testUsers[i])
	}

	//TODO: testing getGroupMembers

	testGroup := models.TestGroup(t)
	err := s.Group().Create(testGroup)
	assert.NoError(t, err)

	for i := range testUsers {
		err := s.Group().AddGroupMember(testUsers[i].ID, testGroup.ID, testUsers[i].ID)
		assert.NoError(t, err)
	}
	membersCount, err := s.Group().GetMembersCount(testGroup.ID)
	assert.NoError(t, err)
	assert.Equal(t, len(testUsers), membersCount)
}
