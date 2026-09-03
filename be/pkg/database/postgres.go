package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"rebutin/config"
)

func NewPostgresDB(cfg *config.Config, log *logrus.Logger) (*sqlx.DB, error) {
	dsn := cfg.GetDSN()
	log.Infof("Connecting to PostgreSQL database at %s:%s...", cfg.DBHost, cfg.DBPort)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Configure Connection Pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	log.Info("Successfully connected to PostgreSQL database")
	return db, nil
}
