package postgres

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

type ContainerLogger struct {
	t *testing.T
}

func (l ContainerLogger) Printf(format string, v ...interface{}) {
	l.t.Logf(format, v...)
}

func GetPostgresTestContainer(
	ctx context.Context,
	migrations []string,
	db, user, password string,
) (*postgres.PostgresContainer, error) {
	return postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15"),
		testcontainers.WithHostPortAccess(5432),
		postgres.WithInitScripts(migrations...),
		postgres.WithDatabase(db),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(15*time.Second)),
	)
}
