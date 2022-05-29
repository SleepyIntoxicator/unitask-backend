package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"context"
	"time"
)

type UniversityRepository struct {
	store *Store
}

func (r *UniversityRepository) Create(ctx context.Context, university *models.University) error {
	query := `INSERT INTO university ( name, location, site, added_at) 
					VALUES ($1, $2, $3, $4) RETURNING id, added_at`
	err := r.store.db.QueryRowContext(ctx, query,
		university.Name,
		university.Location,
		university.Site,
		time.Now()).Scan(&university.ID, &university.AddedAt)
	return err
}

func (r *UniversityRepository) Find(ctx context.Context, id int) (*models.University, error) {
	university := &models.University{}
	query := `SELECT id, name, location, site, added_at FROM university WHERE id = $1`
	err := r.store.db.QueryRowContext(ctx, query, id).Scan(
		&university.ID,
		&university.Name,
		&university.Location,
		&university.Site,
		&university.AddedAt)
	return university, store.HandleErrorNoRows(err)
}
