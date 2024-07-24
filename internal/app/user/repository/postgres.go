package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"github.com/lib/pq"
	authModel "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/user"
	userModel "github.com/raffops/auth/internal/app/user/models"
	"github.com/raffops/chat/pkg/errs"
	"github.com/raffops/chat/pkg/logger"
	"go.uber.org/zap"
	"regexp"
	"slices"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func (p PostgresRepository) GetDB() *sql.DB {
	return p.db
}

// fetchUser is a helper method to create a 'model.User' object
func (p PostgresRepository) fetchUser(
	id, username, email, loginHistory sql.NullString,
	role, status, authType sql.NullInt16,
	createdAt, updatedAt, deleteAt sql.NullTime,
) (userModel.User, errs.ChatError) {

	fetchUser := userModel.User{}
	if id.Valid {
		fetchUser.Id = id.String
	}
	if username.Valid {
		fetchUser.Username = username.String
	}
	if email.Valid {
		fetchUser.Email = email.String
	}
	if role.Valid {
		fetchUser.Role = authModel.RoleId(role.Int16)
	}
	if status.Valid {
		fetchUser.Status = userModel.StatusId(status.Int16)
	}
	if authType.Valid {
		fetchUser.AuthType = userModel.AuthTypeId(authType.Int16)
	}
	if loginHistory.Valid {
		var loginHistoryObj []userModel.LoginHistory
		err := json.Unmarshal([]byte(loginHistory.String), &loginHistoryObj)
		if err != nil {
			return userModel.User{}, errs.NewError(errs.ErrInternal, err)
		}
		fetchUser.LoginHistory = loginHistoryObj
	}

	if createdAt.Valid {
		fetchUser.CreatedAt = createdAt.Time.UTC()
	}

	if updatedAt.Valid {
		fetchUser.UpdatedAt = updatedAt.Time.UTC()
	}
	if deleteAt != (sql.NullTime{}) {
		fetchUser.DeletedAt = deleteAt.Time.UTC()
	}

	return fetchUser, nil

}

// GetUser fetches a userModel from the database. It takes a key and value as arguments
//
// The key is the column name in the database and the value is the value to search for.
// It returns a userModel.User object and an errs.Error, with svcError and appError.
// The error can be a 'errs.ErrNotFound' or 'errs.ErrInternal' error.
// If the userModel is found, the error is nil.
// If the userModel is not found, the svcError is 'errs.ErrNotFound'.
// If there is an internal server error, the svcError is 'errs.ErrInternal'.
// If the key is invalid, the svcError is 'errs.ErrBadRequest'.
func (p PostgresRepository) GetUser(
	ctx context.Context,
	key string,
	value interface{},
) (userModel.User, errs.ChatError) {
	if !slices.Contains([]string{"id", "username", "email"}, key) {
		return userModel.User{}, errs.NewError(errs.ErrBadRequest, errors.New("invalid key. Valid keys are id, username, email"))
	}
	if value == "" {
		return userModel.User{}, errs.NewError(errs.ErrBadRequest, errors.New("invalid value"))
	}
	var id, username, email, loginHistory sql.NullString
	var roleId, statusId, authTypeId sql.NullInt16
	var createdAt, updatedAt, deleteAt sql.NullTime

	sb := buildSelectQuery(key, value)
	queryString, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	err := p.db.QueryRowContext(ctx, queryString, args...).
		Scan(
			&id,
			&username,
			&email,
			&authTypeId,
			&roleId,
			&statusId,
			&loginHistory,
			&createdAt,
			&updatedAt,
			&deleteAt,
		)

	if err != nil {
		return userModel.User{}, getSelectError(err, key, value.(string))
	}
	logger.Debug("User found", zap.String("id", id.String), zap.String("username", username.String))
	return p.fetchUser(
		id,
		username,
		email,
		loginHistory,
		roleId,
		statusId,
		authTypeId,
		createdAt,
		updatedAt,
		deleteAt,
	)
}

func getSelectError(err error, key, value string) errs.ChatError {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errs.NewError(errs.ErrNotFound, fmt.Errorf("user with %s=%s not found", key, value))
	default:
		return errs.NewError(errs.ErrInternal, err)
	}
}

func buildSelectQuery(key string, value interface{}) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id",
		"username",
		"email",
		"auth_type",
		"role",
		"status",
		"login_history",
		"created_at",
		"updated_at",
		"deleted_at",
	).
		From("public.user").
		Where(sb.Equal(key, value))
	return sb
}

// CreateUser inserts a userModel into the database. It takes a userModel.User object as an argument
func (p PostgresRepository) CreateUser(ctx context.Context, tx *sql.Tx, u userModel.User) (userModel.User, errs.ChatError) {
	loginHistory, err := json.Marshal(u.LoginHistory)
	if err != nil {
		return userModel.User{}, errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error marshalling login history: %w", err),
		)
	}

	queryString, args := buildCreateQuery(u, loginHistory)
	var id string
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, queryString, args...).Scan(&id, &createdAt)
	if err != nil {
		return userModel.User{}, getCreateError(err)
	}
	u.Id = id
	u.CreatedAt = createdAt
	return u, nil
}

func getCreateError(err error) errs.ChatError {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" {
			pattern := "Key \\((.*)\\)=\\((.*)\\) already exists."
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(pqErr.Detail)
			if len(matches) == 3 {
				return errs.NewError(
					errs.ErrConflict,
					fmt.Errorf("user with %s=%s already exists", matches[1], matches[2]),
				)
			}
		}
	}
	return errs.NewError(errs.ErrInternal, err)
}

func buildCreateQuery(u userModel.User, loginHistory []byte) (string, []interface{}) {
	sb := sqlbuilder.NewInsertBuilder()
	sb.InsertInto("public.user").
		Cols("username", "email", "auth_type", "role", "status", "login_history").
		Values(u.Username,
			u.Email,
			u.AuthType,
			u.Role,
			u.Status,
			loginHistory,
		)
	queryString, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	queryString += " RETURNING id, created_at"
	return queryString, args
}

// UpdateUser updates a userModel in the database. It takes a userModel.User object and a transaction as arguments
func (p PostgresRepository) UpdateUser(ctx context.Context, tx *sql.Tx, u userModel.User) (
	userModel.User,
	errs.ChatError,
) {
	loginHistory, err := json.Marshal(u.LoginHistory)
	if err != nil {
		return userModel.User{}, errs.NewError(
			errs.ErrInternal,
			fmt.Errorf("error marshalling login history: %w", err),
		)
	}
	loginHistoryStr := sql.NullString{
		String: string(loginHistory),
		Valid:  len(u.LoginHistory) > 0,
	}

	queryString, args := buildUpdateQuery(u, loginHistoryStr)

	var updatedAt sql.NullTime
	err = tx.QueryRowContext(ctx, queryString, args...).Scan(&updatedAt)
	if err != nil {
		return userModel.User{}, getUpdateError(err, u.Id)
	}

	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return u, nil
}

func getUpdateError(err error, id string) errs.ChatError {
	if errors.Is(err, sql.ErrNoRows) {
		return errs.NewError(errs.ErrNotFound, fmt.Errorf("user with id %s not found", id))
	}
	return errs.NewError(errs.ErrInternal, err)
}

func buildUpdateQuery(u userModel.User, loginHistoryStr sql.NullString) (string, []interface{}) {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("public.user").
		Set(
			sb.Assign("username", u.Username),
			sb.Assign("email", u.Email),
			sb.Assign("auth_type", u.AuthType),
			sb.Assign("role", u.Role),
			sb.Assign("status", u.Status),
			sb.Assign("login_history", loginHistoryStr),
		).
		Where(sb.Equal("id", u.Id))
	queryString, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	queryString += " RETURNING updated_at"
	return queryString, args
}

func (p PostgresRepository) DeleteUser(
	ctx context.Context,
	tx *sql.Tx,
	u userModel.User,
) (userModel.User, errs.ChatError) {
	queryString, args := buildDeleteQuery(u)

	var deletedAt sql.NullTime
	err := tx.QueryRowContext(ctx, queryString, args...).Scan(&deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return userModel.User{}, errs.NewError(errs.ErrNotFound, fmt.Errorf("user with id %s not found", u.Id))
	}
	if err != nil {
		return u, errs.NewError(errs.ErrInternal, err)
	}
	if deletedAt.Valid {
		u.DeletedAt = deletedAt.Time
	}
	return u, nil
}

func buildDeleteQuery(u userModel.User) (string, []interface{}) {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("public.user").
		Set(
			sb.Assign("deleted_at", "NOW()"),
			sb.Assign("status", u.Status),
		).
		Where(sb.Equal("id", u.Id))
	queryString, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	queryString += " RETURNING deleted_at"
	return queryString, args
}

// ListUsers fetches a list of users from the database. It takes columns, filters, sorts and pagination as arguments
//
// See 'userModel.ValidColumnsToFetch' for valid columns, 'userModel.ValidColumnsToFilter'
// for valid filters, and 'userModel.ValidColumnsToSort' for valid sorts.
func (p PostgresRepository) ListUsers(
	ctx context.Context,
	columns []string,
	filters []userModel.Filter,
	sorts []userModel.Sort,
	page userModel.Pagination,
) ([]userModel.User, errs.ChatError) {

	queryString, args, err := buildListQuery(columns, filters, sorts, page)
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}

	rows, err := p.db.QueryContext(ctx, queryString, args...)
	if err != nil {
		return nil, errs.NewError(errs.ErrInternal, err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Debug("error closing rows", zap.Error(err))
		}
	}(rows)

	users := make([]userModel.User, 0)
	for rows.Next() {
		var id, username, email, loginHistory sql.NullString
		var roleId, statusId, authTypeId sql.NullInt16
		var createdAt, updatedAt, deleteAt sql.NullTime

		mapColumns := map[string]interface{}{
			"id":            &id,
			"username":      &username,
			"email":         &email,
			"auth_type":     &authTypeId,
			"role":          &roleId,
			"status":        &statusId,
			"login_history": &loginHistory,
			"created_at":    &createdAt,
			"updated_at":    &updatedAt,
			"deleted_at":    &deleteAt,
		}
		columnsToScan := make([]interface{}, 0)
		for _, column := range columns {
			columnsToScan = append(columnsToScan, mapColumns[column])
		}

		err = rows.Scan(columnsToScan...)
		if err != nil {
			return nil, errs.NewError(errs.ErrInternal, err)
		}
		fetchUser, err := p.fetchUser(
			id,
			username,
			email,
			loginHistory,
			roleId,
			statusId,
			authTypeId,
			createdAt,
			updatedAt,
			deleteAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.ErrInternal, err)
		}
		users = append(users, fetchUser)
	}
	return users, nil
}

func buildListQuery(columns []string,
	filters []userModel.Filter,
	sorts []userModel.Sort,
	page userModel.Pagination) (string, []interface{}, error) {

	sb := sqlbuilder.NewSelectBuilder()

	if len(columns) == 0 {
		return "", nil, errs.NewError(errs.ErrBadRequest, errors.New("no columns to fetch"))
	}
	for _, column := range columns {
		if !slices.Contains(userModel.ValidColumnsToFetch, column) {
			return "", nil, errs.NewError(errs.ErrBadRequest, fmt.Errorf("invalid column %s", column))
		}
	}
	sb.Select(columns...).From("public.user")

	for _, filter := range filters {
		if !slices.Contains(userModel.ValidColumnsToFilter, filter.Key) {
			return "", nil, errs.NewError(errs.ErrBadRequest, errors.New("invalid filter key"))
		}
		switch filter.Comparison {
		case userModel.ComparisonEqual:
			sb.Where(sb.Equal(filter.Key, filter.Value))
		case userModel.ComparisonGreaterThan:
			sb.Where(sb.GreaterThan(filter.Key, filter.Value))
		case userModel.ComparisonGreaterThanOrEqual:
			sb.Where(sb.GreaterEqualThan(filter.Key, filter.Value))
		case userModel.ComparisonLessThan:
			sb.Where(sb.LessThan(filter.Key, filter.Value))
		case userModel.ComparisonLessThanOrEqual:
			sb.Where(sb.LessEqualThan(filter.Key, filter.Value))
		default:
			return "", nil, errs.NewError(errs.ErrBadRequest, errors.New("invalid comparison operator"))
		}
	}

	if len(sorts) == 0 {
		sb.OrderBy("created_at").Desc()
	}
	for _, sort := range sorts {
		if !slices.Contains(userModel.ValidColumnsToSort, sort.Key) {
			return "", nil, errs.NewError(errs.ErrBadRequest, errors.New("invalid sort key"))
		}
		if sort.Order == userModel.OrderAsc {
			sb.OrderBy(sort.Key).Asc()
		} else {
			sb.OrderBy(sort.Key).Desc()
		}
	}

	sb.Offset(page.Offset).Limit(page.Limit)

	s, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	return s, args, nil
}

func NewPostgresUserRepository(db *sql.DB) user.ReaderWriterRepository {
	return &PostgresRepository{db: db}
}
