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
	Cfg       *config.Config
	DB        *sql.DB
	Queue     UseQueue
	StartTime time.Time
}

func newDB(cfg *config.Config) *dbStorage {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		log.Fatalln("Failed open DB on startup: ", err)
	}
	if err = makeMigrate(db); err != nil {
		log.Fatalln("Failed migrate DB: ", err)
	}
	storage := &dbStorage{
		Cfg:       cfg,
		Queue:     newQueue(),
		StartTime: time.Now(),
		DB:        db,
	}

	go func() {
		orders, err := storage.GetOrdersForUpdate(context.TODO())
		if err != nil {
			log.Println("Failed get orders for update")
			return
		}
		if len(orders) == 0 {
			log.Println("Updated all old orders")
			return
		}
		if err = storage.Queue.Push(orders); err != nil {
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
