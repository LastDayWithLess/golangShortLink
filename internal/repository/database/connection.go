package database

import (
	"database/sql"
	"errors"
	"fmt"
	"short_link/config"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrOpenDB = errors.New("failed to open database")
	ErrConnDB = errors.New("failed to connect to database")
)

type ConectionPool struct {
	db *sql.DB
}

func NewConnection(cfg config.DataBaseConfig) (*ConectionPool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, ErrOpenDB
	}

	if err := db.Ping(); err != nil {
		return nil, ErrConnDB
	}

	configurePool(db)

	return &ConectionPool{db: db}, nil
}

func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)
}

func (cp *ConectionPool) GetDB() *sql.DB {
	if cp == nil {
		return nil
	}
	return cp.db
}

func (cp *ConectionPool) CloseDB() error {
	if cp.db != nil {
		return cp.db.Close()
	}

	return nil
}
