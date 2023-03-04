package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/pkg"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type dbStorage struct {
	cfg       *config.Config
	db        *sql.DB
	queue     UseQueue
	startTime time.Time
}

func newDBStorage(cfg *config.Config) *dbStorage {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		pkg.Log.Fatal(err)
	}
	if err = makeMigrate(db); err != nil {
		pkg.Log.Fatal(err)
	}
	storage := &dbStorage{
		cfg:       cfg,
		queue:     newQueue(),
		startTime: time.Now(),
		db:        db,
	}

	go func() {
		orders, err := storage.GetOrdersForUpdate(context.TODO())
		if err != nil {
			return
		}
		if len(orders) == 0 {
			return
		}
		if err = storage.queue.Push(orders); err != nil {
			pkg.Log.Info("Failed push orders to queue")
			return
		}
	}()
	return storage
}

func makeMigrate(db *sql.DB) error {
	/*
		Creating DB tables.
	*/
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		pkg.Log.Err(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		pkg.Log.Err(err)
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		pkg.Log.Fatal(err)
		return err
	}
	return nil
}
