package utils

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"
	"github.com/tespkg/bytes-be/res_embed"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/url"
	"strings"
)

func Migrate(dsn string) error {
	if err := createDatabaseIfNotExists(dsn); err != nil {
		return err
	}

	fs, err := iofs.New(res_embed.PgMigrationFiles, "migration/pg")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", fs, dsn)
	if err != nil {
		return err
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func createDatabaseIfNotExists(dsn string) error {
	dsnWithoutDbName, err := getDsnWithoutDbName(dsn)
	if err != nil {
		return err
	}

	dbName, err := getDbNameFromDsn(dsn)
	if err != nil {
		return err
	}

	// create the database if it does not exist
	db, err := gorm.Open(postgres.Open(dsnWithoutDbName), &gorm.Config{})
	if err != nil {
		return err
	}

	var count int64
	if err = db.
		Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", dbName).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		sqlStr := fmt.Sprintf(`CREATE DATABASE "%s"`, dbName)
		if err = db.Exec(sqlStr).Error; err != nil {
			return err
		}
	}

	return nil
}

func getDsnWithoutDbName(dsn string) (string, error) {
	parsedUrl, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = ""
	return parsedUrl.String(), nil
}

func getDbNameFromDsn(dsn string) (string, error) {
	parsedUrl, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(parsedUrl.Path, "/"), nil
}
