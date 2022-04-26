package teststore

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
)

type Store struct {
	authRepository       *AuthRepository
	userRepository       *UserRepository
	taskRepository       *TaskRepository
	universityRepository *UniversityRepository
	groupRepository      *GroupRepository
	subjectRepository    *SubjectRepository
}

func New() *Store {
	return &Store{}
}

func (s *Store) Auth() store.AuthRepository {
	if s.authRepository == nil {
		s.authRepository = &AuthRepository{
			store: s,
		}
	}

	return s.authRepository
}

func (s *Store) User() store.UserRepository {
	if s.userRepository == nil {
		s.userRepository = &UserRepository{
			store: s,
			users: make(map[int]*models.User),
		}
	}

	return s.userRepository
}

func (s *Store) Task() store.TaskRepository {
	if s.taskRepository == nil {
		s.taskRepository = &TaskRepository{
			store: s,
		}
	}
	//TODO: implement methods
	return s.taskRepository
}

func (s *Store) University() store.UniversityRepository {
	if s.universityRepository == nil {
		s.universityRepository = &UniversityRepository{
			store: s,
		}
	}
	return s.universityRepository
}

func (s *Store) Group() store.GroupRepository {
	if s.groupRepository == nil {
		s.groupRepository = &GroupRepository{
			store: s,
		}
	}

	return s.groupRepository
}

func (s *Store) Subject() store.SubjectRepository {
	if s.subjectRepository == nil {
		s.subjectRepository = &SubjectRepository{
			store: s,
		}
	}
	//TODO: implement methods
	return s.subjectRepository
}
