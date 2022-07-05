package migrations

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
)

type Config struct {
	Enable    bool   `mapstructure:"enable"`
	Recreate  bool   `mapstructure:"recreate"`
	SourceURL string `mapstructure:"sourceURL" validate:"required,gte=0"`
	DbURL     string `mapstructure:"dbURL" validate:"required,gte=0"`
}

func RunMigrations(cfg Config) (version uint, dirty bool, err error) {
	if !cfg.Enable {
		return 0, false, nil
	}

	m, err := migrate.New(cfg.SourceURL, cfg.DbURL)
	if err != nil {
		return 0, false, err
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			err = sourceErr
		}
		if dbErr != nil {
			err = dbErr
		}
	}()

	if cfg.Recreate {
		if err := m.Down(); err != nil {
			return 0, false, err
		}
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 0, false, err
	}

	return m.Version()
}
