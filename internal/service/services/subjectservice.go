package services

import (
	"backend/internal/api/v1/models"
	"backend/internal/service"
	"backend/internal/store"
)

type SubjectService struct {
	service *Service
}

func (s *SubjectService) Create(subject *models.Subject) error {
	if err := subject.Validate(); err != nil {
		return err
	}

	if err := s.service.store.Subject().Create(subject); err != nil {
		return err
	}

	return nil
}

func (s *SubjectService) GetAllSubjects(limit, offset int) ([]models.Subject, error) {
	subjects, err := s.service.store.Subject().GetAll(limit, offset)
	if err != nil && err != store.ErrRecordNotFound {
		return nil, err
	} else {
		return subjects, nil
	}

}

func (s *SubjectService) Find(subjectID int) (*models.Subject, error) {
	subject, err := s.service.store.Subject().Find(subjectID)
	if err != nil && err == store.ErrRecordNotFound {
		return nil, service.ErrSubjectNotFound
	} else if err != nil {
		return nil, err
	}

	return subject, nil
}

func (s *SubjectService) Delete(subjectID int) (*models.Subject, error) {
	subject, err := s.service.store.Subject().Delete(subjectID)
	if err != nil {
		return nil, err
	}

	return subject, nil
}
