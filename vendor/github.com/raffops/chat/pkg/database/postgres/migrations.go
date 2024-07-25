package postgres

import (
	"os"
	"path/filepath"
	"strings"
)

func GetMigrations() ([]string, error) {
	migrationsPath := filepath.Join("internal", "database", "postgres", "migrations")
	migrations, err := os.ReadDir(migrationsPath)
	if err != nil {
		return nil, err
	}
	var migrationFiles []string
	for _, migration := range migrations {
		if strings.HasSuffix(migration.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsPath, migration.Name()))
		}
	}
	return migrationFiles, nil
}
