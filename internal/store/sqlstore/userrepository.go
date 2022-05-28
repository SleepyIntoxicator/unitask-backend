package sqlstore

import (
	"backend/internal/api/v1/models"
	"backend/internal/store"
	"context"
	"database/sql"
	"time"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := user.Validate(); err != nil {
		return err
	}

	if err := user.BeforeCreate(); err != nil {
		return err
	}

	query := `INSERT INTO public.user (login, full_name, email, encrypted_password, created_at) 
						  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.store.db.QueryRowContext(ctx, query,
		user.Login,
		user.FullName,
		user.Email,
		user.EncryptedPassword,
		time.Now(),
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}

func (r *UserRepository) GetAll(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM public.user ORDER BY id`
	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	rows, err := r.store.db.QueryContext(ctx, query)
	if err != nil {
		return nil, store.HandleErrorNoRows(err)
	}

	for rows.Next() {
		u := models.User{}

		if err := rows.Scan(
			&u.ID,
			&u.Login,
			&u.FullName,
			&u.Email,
			&u.EncryptedPassword,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}
	if nil == users {
		return nil, store.ErrNoRowsFound
	}
	return users, nil
}

func (r *UserRepository) Find(ctx context.Context, userID int) (*models.User, error) {
	u := &models.User{}

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE id = $1`
	if err := r.store.db.QueryRowContext(ctx, query, userID).Scan(
		&u.ID,
		&u.Login,
		&u.FullName,
		&u.Email,
		&u.EncryptedPassword,
		&u.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE login = $1`
	if err := r.store.db.QueryRowContext(ctx, query, login).Scan(
		&u.ID,
		&u.Login,
		&u.FullName,
		&u.Email,
		&u.EncryptedPassword,
		&u.CreatedAt,
	); err != nil {
		return nil, store.HandleErrorNoRows(err)
	}

	return u, nil
}

// FindByEmail Returns store.ErrRecordNotFound if db driver returned sql.ErrNoRows
//or another error if unknown one occurred.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE email = $1`
	if err := r.store.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Login,
		&u.FullName,
		&u.Email,
		&u.EncryptedPassword,
		&u.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return u, nil
}

// CreateTester Only for testing
func (r *UserRepository) CreateTester() (*models.User, error) {
	u := &models.User{
		Login:     "tester",
		FullName:  "User Test",
		Email:     "tester@example.org",
		Password:  "testerPassword",
		CreatedAt: time.Now(),
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	okU := &models.User{}
	query := `select id, login, full_name, email, encrypted_password, created_at from public.user where login = 'tester'`
	if err := r.store.db.QueryRow(query).Scan(
		&okU.ID,
		&okU.Login,
		&okU.FullName,
		&okU.Email,
		&okU.EncryptedPassword,
		&okU.CreatedAt,
	); err == sql.ErrNoRows {
		query := `INSERT INTO public.user (login, full_name, email, encrypted_password) 
								   VALUES ($1, $2, $3, $4) RETURNING id`
		err = r.store.db.QueryRow(query,
			u.Login,
			u.FullName,
			u.Email,
			u.EncryptedPassword,
		).Scan(
			&u.ID,
		)
		return u, err
	} else if err != nil {
		return nil, err
	}
	return okU, nil
}

func (r *UserRepository) IsUserExist(ctx context.Context, userID int) (bool, error) {
	query := `SELECT FROM public.user WHERE id = $1`
	_, err := r.store.db.ExecContext(ctx, query, userID)
	return store.HandleIsFieldFounded(err)
}

func (r *UserRepository) GetUserRoles(ctx context.Context, userID int) ([]models.Role, error) {
	var roles []models.Role

	err := r.store.db.SelectContext(ctx, &roles,
		`SELECT id, name, description FROM public.role
                             WHERE id in (SELECT user_id FROM userroles WHERE user_id = $1)`,
		userID)
	return roles, store.HandleErrorNoRows(err)
}

func (r *UserRepository) GetUserToken(ctx context.Context, userID int) (*models.UserToken, error) {
	t := &models.UserToken{}

	query := `SELECT * FROM usertoken where user_id = $1`
	err := r.store.db.SelectContext(ctx, t, query, userID)

	return t, err
}
