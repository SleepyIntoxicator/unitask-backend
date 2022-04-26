package sqlstore

import (
	"back-end/internal/app/api/v1/models"
	"back-end/internal/app/store"
	"database/sql"
	"time"
)

type UserRepository struct {
	store *Store
}

func (r *UserRepository) Create(u *models.User) error {
	if err := u.Validate(); err != nil {
		return err
	}

	if err := u.BeforeCreate(); err != nil {
		return err
	}

	query := `INSERT INTO public.user (login, full_name, email, encrypted_password, created_at) 
						  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.store.db.QueryRow(query,
		u.Login,
		u.FullName,
		u.Email,
		u.EncryptedPassword,
		time.Now(),
	).Scan(
		&u.ID,
		&u.CreatedAt,
	)
}

func (r *UserRepository) GetAll(limit, offset int) ([]models.User, error) {
	var users []models.User

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM public.user ORDER BY id`
	query, err := r.store.AddLimitAndOffsetToQuery(query, limit, offset)
	if err != nil {
		return nil, err
	}

	rows, err := r.store.db.Query(query)
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

func (r *UserRepository) FindByLogin(login string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE login = $1`
	if err := r.store.db.QueryRow(query, login).Scan(
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

//Returns store.ErrRecordNotFound if db driver returned sql.ErrNoRows
//or another error if unknown one occurred.
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	u := &models.User{}

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE email = $1`
	if err := r.store.db.QueryRow(query, email).Scan(
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

func (r *UserRepository) Find(id int) (*models.User, error) {
	u := &models.User{}

	query := `SELECT id, login, full_name, email, encrypted_password, created_at FROM "user" WHERE id = $1`
	if err := r.store.db.QueryRow(query, id).Scan(
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

// Only for testing
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

func (r *UserRepository) IsUserExist(userID int) (bool, error) {
	query := `SELECT FROM public.user WHERE id = $1`
	_, err := r.store.db.Exec(query, userID)
	return store.HandleIsFieldFounded(err)
}

func (r *UserRepository) GetUserRoles(userID int) ([]models.Role, error) {
	var roles []models.Role

	err := r.store.db.Select(&roles,
		"SELECT id, name, description FROM public.role WHERE id in (SELECT user_id FROM userroles WHERE user_id = $1)",
		userID)
	return roles, store.HandleErrorNoRows(err)
}

func (r *UserRepository) GetUserToken(userID int) (*models.UserToken, error) {
	t := &models.UserToken{}

	query := `SELECT * FROM usertoken where user_id = $1`
	err := r.store.db.Select(t, query, userID)

	return t, err
}
