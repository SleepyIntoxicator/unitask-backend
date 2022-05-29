package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
	"context"
)

type SubjectService struct {
	service *Service
}

func (s *SubjectService) Create(ctx context.Context, subject *models.Subject) error {
	if err := subject.Validate(); err != nil {
		return err
	}

	if err := s.service.store.Subject().Create(ctx, subject); err != nil {
		return err
	}

	return nil
}

func (s *SubjectService) GetAllSubjects(ctx context.Context, limit, offset int) ([]models.Subject, error) {
	subjects, err := s.service.store.Subject().GetAll(ctx, limit, offset)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else {
		return subjects, nil
	}

}

func (s *SubjectService) Find(ctx context.Context, subjectID int) (*models.Subject, error) {
	subject, err := s.service.store.Subject().Find(ctx, subjectID)
	if err != nil && err == store.ErrRecordNotFound {
		return nil, service.ErrSubjectNotFound
	} else if err != nil {
		return nil, err
	}

	return subject, nil
}

func (s *SubjectService) Delete(ctx context.Context, subjectID int) (*models.Subject, error) {
	subject, err := s.service.store.Subject().Delete(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	return subject, nil
}
