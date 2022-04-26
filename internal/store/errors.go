package store

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/sirupsen/logrus"
)

var (
	//ErrRecordNotFound For one record
	ErrRecordNotFound = errors.New("record not found")
	//ErrNoRowsFound For multiple records
	ErrNoRowsFound = errors.New("the query returned an empty response")

	ErrUserNotFound  = errors.New("user not found")
	ErrGroupNotFound = errors.New("group not found")
	//ErrLink
)

// HandleErrorNoRows returns ErrRecordNotFound if error is equal sql.ErrNoRows
// Returns other error, if not
func HandleErrorNoRows(err error) error {
	if err == sql.ErrNoRows {
		return ErrRecordNotFound
	}
	return err
}

func HandleIgnoreErrorNoRows(err error) error {
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

//HandleIsFieldFounded return true if the field was founded or false if not.
//	If an error occurs during the execution of the method, the method returns error.
func HandleIsFieldFounded(err error) (bool, error) {
	//ignore sql.ErrNoRows
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func HandlePgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.Is(err, pgErr) {
		pgErr = err.(*pgconn.PgError)
		logrus.WithFields(logrus.Fields{
			"error":    pgErr.Message,
			"detail":   pgErr.Detail,
			"where":    pgErr.Where,
			"code":     pgErr.Code,
			"SQLState": pgErr.SQLState(),
		}).Error(pgErr.Error())
		Err := fmt.Errorf(
			"SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
			pgErr.Message,
			pgErr.Detail,
			pgErr.Where,
			pgErr.Code,
			pgErr.SQLState())
		return Err
	}
	return err
}
