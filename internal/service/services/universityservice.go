package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
)

type UniversityService struct {
	service *Service
}

func (s *UniversityService) Create(ctx context.Context, university *models.University) error {
	err := s.service.store.University().Create(ctx, university)
	if err != nil {
		return err
	}
	return err
}

func (s *UniversityService) Find(ctx context.Context, id int) (*models.University, error) {
	university, err := s.service.store.University().Find(ctx, id)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else if err == store.ErrRecordNotFound {
		return nil, service.ErrUniversityNotFound
	}
	return university, nil
}
