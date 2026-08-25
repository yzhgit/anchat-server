package database

import (
	"flamingo/pkg/config"
	"fmt"
	"net/url"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

var ProviderSet = wire.NewSet(NewDB)

// NewDB creates a new database connection and returns a cleanup function
// that closes the underlying sql.DB. The func() return satisfies Wire's
// cleanup convention: when a provider returns (T, func(), error), Wire
// calls cleanup functions in reverse order on error and aggregates them
// into a single func() returned from the injector.
func NewDB(cfg config.Database) (*gorm.DB, func(), error) {
	encodedPassword := url.QueryEscape(cfg.Password)

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User,
		encodedPassword,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	if cfg.SSLMode != "" {
		dsn += "?sslmode=" + cfg.SSLMode
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	if err := db.Use(tracing.NewPlugin(
		tracing.WithTracerProvider(otel.GetTracerProvider()),
		tracing.WithAttributes(semconv.DBSystemPostgreSQL))); err != nil {
		// Attempt to close the already-opened sql.DB on failure
		sqlDB, sqlErr := db.DB()
		if sqlErr == nil {
			_ = sqlDB.Close()
		}
		return nil, nil, err
	}

	return db, func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Errorf("cleanup: get sql.DB failed: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Errorf("cleanup: close database failed: %v", err)
		}
	}, nil
}
