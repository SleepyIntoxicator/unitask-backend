package services

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/service"
	"back-end/internal/app/store"
)

type UniversityService struct {
	service *Service
}

func (s *UniversityService) Create(university *models.University) error {
	err := s.service.store.University().Create(university)
	if err != nil {
		return err
	}
	return err
}

func (s *UniversityService) Find(id int) (*models.University, error) {
	university, err := s.service.store.University().Find(id)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else if err == store.ErrRecordNotFound {
		return nil, service.ErrUniversityNotFound
	}
	return university, nil
}