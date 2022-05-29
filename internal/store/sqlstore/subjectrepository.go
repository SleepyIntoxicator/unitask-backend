package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"context"
	"database/sql"
)

type SubjectRepository struct {
	store *Store
}

func (r *SubjectRepository) Create(ctx context.Context, subject *models.Subject) error {
	return r.store.db.QueryRowContext(ctx,
		"INSERT INTO subject (name) values ($1) RETURNING id",
		subject.Name,
	).Scan(&subject.ID)
}

func (r *SubjectRepository) GetAll(ctx context.Context, limit, offset int) ([]models.Subject, error) {
	var subjects []models.Subject
	query := `SELECT id, name FROM subject ORDER BY id`

	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	err = r.store.db.SelectContext(ctx, &subjects, query)
	if err != nil {
		return nil, err
	}

	return subjects, nil
}

func (r *SubjectRepository) Find(ctx context.Context, id int) (*models.Subject, error) {
	s := &models.Subject{}

	if err := r.store.db.QueryRowContext(ctx,
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

func (r *SubjectRepository) Delete(ctx context.Context, id int) (*models.Subject, error) {
	subject := &models.Subject{}
	err := r.store.db.QueryRowContext(ctx, "DELETE FROM subject WHERE id = $1 RETURNING id, name", id).Scan(
		&subject.ID,
		&subject.Name)
	return subject, store.HandleErrorNoRows(err)
}
