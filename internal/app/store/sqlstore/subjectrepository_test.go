package sqlstore

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubjectRepository_Create(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("subject")

	subj := models.TestSubject(t)


	s := New(db)
	assert.NoError(t, s.Subject().Create(subj))
	assert.NotNil(t, subj)
}

func TestSubjectRepository_Find(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("subject")

	subj := models.TestSubject(t)

	s := New(db)
	_, err := s.Subject().Find(subj.ID)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())

	assert.NoError(t, s.Subject().Create(subj))
	subj2, err := s.Subject().Find(subj.ID)
	assert.NoError(t, err)
	assert.NotNil(t, subj2)
}

func TestSubjectRepository_Delete(t *testing.T) {
	db, teardown := TestDB(t, databaseDriver, databaseURL)
	defer teardown("subject")

	s := New(db)
	subj := models.TestSubject(t)

	assert.NoError(t, s.Subject().Create(subj))
	assert.NoError(t, s.Subject().Delete(subj.ID))
	_, err := s.Subject().Find(subj.ID)
	assert.EqualError(t, err, store.ErrRecordNotFound.Error())
}