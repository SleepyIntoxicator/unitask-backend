package apiserver

import (
	cfg "back-end/internal/app/config"
	"back-end/internal/app/store/sqlstore"
	"back-end/pkg/database/postgres"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(configPath string) {
	var httpServer *http.Server

	config, err := cfg.Init(configPath)
	if err != nil {
		log.Error("Unable to init new config: ", err)
		return
	}

	db, err := postgres.NewPostgresConnection(postgres.ConnectionData{
		Host:     config.Postgres.Host,
		Username: config.Postgres.Username,
		Password: config.Postgres.Password,
		DBName:   config.Postgres.DBName,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to stop database conn: %v", err.Error())
		} else {
			log.Printf("database connection is stoped")
		}
	}()

	store := sqlstore.New(db)
	srv := newServer(store, config)

	httpServer = &http.Server{
		Addr:         config.HTTP.Host + ":" + config.HTTP.Port,
		Handler:      srv,
		ReadTimeout:  config.HTTP.ReadTimeout,
		WriteTimeout: config.HTTP.WriteTimeout,
	}

	// Starting server

	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to start server: %v", err.Error())
		}
	}()
	log.Printf("Server started")

	//Graceful shutdown and waiting for text commands

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Kill, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		var cmd string
		for {
			if _, err := fmt.Scanf("%s", &cmd); err != nil {
				continue
			}
			switch cmd {
			case "stop", "close", "exit", "cancel":
				log.Info("exit call")
				quit <- syscall.SIGTERM
				return
			}
		}
	}()
	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown HTTP server: %v", err)
	} else {
		log.Printf("HTTP server shutdowned")
	}
}

//newDB deprecated
func newDB(databaseDriver, databaseURL string) (*sqlx.DB, error) {
	db, err := sqlx.Open(databaseDriver, databaseURL)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(0) //TODO: check errors at 800>RPS and working longer than 3m
	db.SetConnMaxLifetime(0) //TODO: check errors at 800>RPS and working longer than 3m
	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	log.Info("Connection to the database has started. Timeout of the conn: 5 sec")
	if err := db.Ping(); err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgErr) {
			pgErr = err.(*pgconn.PgError)
			log.WithFields(log.Fields{
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
			return nil, Err
		}

		log.Warn("The ip address of the database server may have been entered incorrectly.")
		return nil, err
	}
	log.Info("Connection to the database was established successfully")

	return db, nil
}
