package teststore

import (
	"backend/internal/api/v1/models"
)

type UniversityRepository struct {
	store *Store
}

func (r *UniversityRepository) Create(university *models.University) error {
	panic("implement me")
}

func (r *UniversityRepository) Find(universityID int) (*models.University, error) {
	panic("implement me")
}
