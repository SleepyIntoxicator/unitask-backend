package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

const connectionTimeout = 5 * time.Second
const maxAttempts = 5

type ConnectionData struct {
	Host     string
	Username string
	Password string
	DBName   string
}

func NewPostgresConnection(connData ConnectionData) (*sqlx.DB, error) {
	if connData.Username == "" || connData.Password == "" {
		return nil, errors.New("invalid connection credentials")
	}

	connTimeout := int(connectionTimeout.Seconds())
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:5432/%s?connect_timeout=%d",
		connData.Username,
		connData.Password,
		connData.Host,
		connData.DBName,
		connTimeout,
	)
	maxWaitingTime := maxAttempts * connectionTimeout

	logrus.Infof("Connection to the database has started. Timeout of the conn: %s", maxWaitingTime)

	var db *sqlx.DB
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), maxWaitingTime)
	defer cancel()
	for attempt := 0; attempt < maxAttempts; attempt++ {
		logrus.Infof("Connection attempt [%d]", attempt+1)
		if db, err = sqlx.ConnectContext(ctx, "pgx", dbURL); err == nil {
			break
		}

		time.Sleep(connectionTimeout)
	}
	if err != nil {
		return nil, err
	}

	logrus.Info("Connection to the database was established successfully")
	return db, nil
}
