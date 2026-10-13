package infra

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kelseyhightower/envconfig"
)

type PostgresConfig struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"postgres"`
	Password string `envconfig:"DB_PASSWORD" default:"postgres"`
	Name     string `envconfig:"DB_NAME" default:"gdctest"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

type PostgresDB struct {
	DB *sql.DB
}

func (p *PostgresDB) Shutdown(ctx context.Context) error {
	if p.DB != nil {
		return p.DB.Close()
	}
	return nil
}

// GetDB returns typed *sql.DB safely from Container
func GetDB(c *Container) (*sql.DB, bool) {
	if c == nil {
		return nil, false
	}
	if res, ok := c.Get("postgres"); ok {
		if pg, ok := res.(*PostgresDB); ok {
			return pg.DB, true
		}
	}
	return nil, false
}

func init() {
	Register("postgres", func(ctx context.Context, c *Container) error {
		var cfg PostgresConfig
		if err := envconfig.Process("", &cfg); err != nil {
			return fmt.Errorf("failed to parse postgres config: %w", err)
		}

		db, err := sql.Open("pgx", cfg.DSN())
		if err != nil {
			return fmt.Errorf("failed to open postgres: %w", err)
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(15 * time.Minute)

		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := db.PingContext(pingCtx); err != nil {
			_ = db.Close()
			return fmt.Errorf("failed to ping postgres: %w", err)
		}

		c.Set("postgres", &PostgresDB{DB: db})
		return nil
	})
}
