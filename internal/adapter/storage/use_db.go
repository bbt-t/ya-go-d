package storage

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/bbt-t/ya-go-d/internal/config"

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
		log.Fatalln("Failed open DB on startup: ", err)
	}
	if err = makeMigrate(db); err != nil {
		log.Fatalln("Failed migrate DB: ", err)
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
			log.Println("Failed push orders to queue")
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
		log.Printf("Failed create postgres instance: %v\n", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver)
	if err != nil {
		log.Printf("Failed create migration instance: %v\n", err)
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Failed migrate: ", err)
		return err
	}
	return nil
}
