package teststore

import (
	"backend/internal/api/v1/models"
)

type SubjectRepository struct {
	store *Store
}

func (r *SubjectRepository) Create(subject *models.Subject) error {
	panic("implement me")
}

func (r *SubjectRepository) GetAll(limit, offset int) ([]models.Subject, error) {
	panic("implement me")
}

func (r *SubjectRepository) Find(i int) (*models.Subject, error) {
	panic("implement me")
}

func (r *SubjectRepository) Delete(i int) (*models.Subject, error) {
	panic("implement me")
}
