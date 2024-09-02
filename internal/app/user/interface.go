package user

import (
	"context"
	"database/sql"

	userModels "github.com/raffops/chat_auth/internal/app/user/models"
	"github.com/raffops/chat_commons/pkg/errs"
)

type ReaderRepository interface {
	GetUser(ctx context.Context, key string, value interface{}) (userModels.User, errs.ChatError)
	ListUsers(
		ctx context.Context,
		columns []string,
		filters []userModels.Filter,
		sorts []userModels.Sort,
		page userModels.Pagination,
	) ([]userModels.User, errs.ChatError)
}

type WriterRepository interface {
	CreateUser(ctx context.Context, tx *sql.Tx, u userModels.User) (userModels.User, errs.ChatError)
	UpdateUser(ctx context.Context, tx *sql.Tx, u userModels.User) (userModels.User, errs.ChatError)
	DeleteUser(ctx context.Context, tx *sql.Tx, u userModels.User) (userModels.User, errs.ChatError)
	GetDB() *sql.DB
}

type ReaderWriterRepository interface {
	ReaderRepository
	WriterRepository
}
