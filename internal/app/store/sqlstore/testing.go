package sqlstore

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"testing"
)

func TestDB(t *testing.T, databaseDriver, databaseURL string) (*sqlx.DB, func (...string)) {
	t.Helper()

	db, err := sqlx.Open(databaseDriver, databaseURL)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	return db, func(tables ...string) {
		if len(tables) > 0 {
			query := fmt.Sprintf("TRUNCATE public.%s CASCADE", strings.Join(tables, ", public."))
			//if _, err := db.Exec(fmt.Sprintf("TRUNCATE public.%s CASCADE", strings.Join(tables, ", public.")));
			if _, err := db.Exec(query);
			err != nil {
				t.Fatal(err)
			}
			for _, table := range tables {
				_, err := db.Exec(fmt.Sprintf("alter sequence if exists %s_id_seq restart", table))
				if err != nil {
					t.Fatal(err)
				}
			}
		}

		db.Close()
	}
}