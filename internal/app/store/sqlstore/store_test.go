package sqlstore

import (
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

var (
	databaseDriver string
	databaseURL string
	logLevel string
)

func TestMain(m *testing.M) {
	databaseDriver = os.Getenv("DATABASE_DRIVER")
	if databaseDriver == "" {
		logrus.Warn("Env value was not found. Default driver name applied in store_test")
		databaseDriver = "pgx"
	}

	databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logrus.Warn("Env value was not found. Default url applied in store_test")
		databaseURL = "host=10.0.0.24 user=postgres password=postgres dbname=apiserver_unitask sslmode=disable"
	}

	logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logrus.Warn("Env value was not found. Default log_level applied in file store_test.go")
		logLevel = "Info"
	}
	//TODO: remove databaseURL from here. Update URL there. Tag: databaseURL, URL, env, dbname

	os.Exit(m.Run())
}