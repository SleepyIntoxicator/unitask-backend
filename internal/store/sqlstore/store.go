package sqlstore

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

type Store struct {
	db *sqlx.DB

	authRepository       *AuthRepository
	userRepository       *UserRepository
	taskRepository       *TaskRepository
	universityRepository *UniversityRepository
	groupRepository      *GroupRepository
	subjectRepository    *SubjectRepository
}

func New(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
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
	return s.subjectRepository
}

func (s *Store) AddLimitAndOffsetToQuery(q string, limit, offset int) (string, error) {
	if limit < 0 || offset < 0 {
		return "", models.ErrLimitOrOffsetLessThanZero
	}

	var lim string
	if limit == 0 {
		lim = "ALL"
	} else {
		lim = fmt.Sprintf("%d", limit)
	}
	return fmt.Sprintf(q+" LIMIT %s OFFSET %d", lim, offset), nil
}
