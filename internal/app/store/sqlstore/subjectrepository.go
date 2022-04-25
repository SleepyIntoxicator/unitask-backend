package sqlstore

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
	"database/sql"
)

type SubjectRepository struct {
	store *Store
}

func (r *SubjectRepository) Create(s *models.Subject) error {
	return r.store.db.QueryRow(
		"INSERT INTO subject (name) values ($1) RETURNING id",
		s.Name,
	).Scan(&s.ID)
}

func (r *SubjectRepository) GetAll(limit, offset int) ([]models.Subject, error) {
	var subjects []models.Subject
	query := `SELECT id, name FROM subject ORDER BY id`

	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	err = r.store.db.Select(&subjects, query)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

func (r *SubjectRepository) Find(id int) (*models.Subject, error) {
	s := &models.Subject{}

	if err := r.store.db.QueryRow(
		"SELECT id, name FROM subject WHERE id = $1",
		id,
	).Scan(
		&s.ID,
		&s.Name,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}
		return nil, err
	}

	return s, nil
}

func (r *SubjectRepository) Delete(id int) (*models.Subject, error) {
	subject := &models.Subject{}
	err := r.store.db.QueryRow("DELETE FROM subject WHERE id = $1 RETURNING id, name", id).Scan(
		&subject.ID,
		&subject.Name)
	return subject, store.HandleErrorNoRows(err)
}
