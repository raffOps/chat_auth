package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	auth "github.com/raffops/auth/internal/app/auth/model"
	user "github.com/raffops/auth/internal/app/user/models"
	"github.com/raffops/auth/internal/app/user/repository/migrations"
	database "github.com/raffops/chat/pkg/database/postgres"
	"github.com/raffops/chat/pkg/errs"
	"log"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

var (
	RandomTimeStr = "2024-05-07T12:27:58Z"
	RandomTime    = time.Date(2024, time.May, 7, 12, 27, 58, 0, time.UTC)
	UserJonhDoe   = user.User{
		Id:        "ac554921-1b75-43bd-9e1d-e17dfb38f6c3",
		Username:  "jon",
		Email:     "john@doe",
		AuthType:  user.AuthTypeGoogle,
		Role:      auth.RoleUser,
		Status:    user.StatusActive,
		CreatedAt: RandomTime,
		UpdatedAt: RandomTime,
	}
	UserJonhDoeUpdated = user.User{
		Id:        "ac554921-1b75-43bd-9e1d-e17dfb38f6c3",
		Username:  "jon",
		Email:     "john@doe",
		AuthType:  user.AuthTypeGithub,
		Role:      auth.RoleUser,
		Status:    user.StatusActive,
		CreatedAt: RandomTime,
		UpdatedAt: RandomTime,
	}
	UserJohnnDoe = user.User{
		Id:        "ac554921-1b75-43bd-9e1d-e17dfb38f6c3",
		Username:  "johnn",
		Email:     "john@doe",
		AuthType:  user.AuthTypeGithub,
		Role:      auth.RoleUser,
		Status:    user.StatusActive,
		CreatedAt: RandomTime,
		UpdatedAt: RandomTime,
	}
	UserJaneDoe = user.User{
		Id:        "b0a860ee-35ac-478a-8961-069fe2b8dfc1",
		Username:  "jane",
		Email:     "jane@doe",
		AuthType:  user.AuthTypeGoogle,
		Role:      auth.RoleAdmin,
		Status:    user.StatusActive,
		CreatedAt: RandomTime,
		UpdatedAt: RandomTime,
	}
	UserTomDoe = user.User{
		Username: "tom",
		Email:    "tom@doe",
		AuthType: user.AuthTypeGoogle,
		Role:     auth.RoleAdmin,
		Status:   user.StatusActive,
		LoginHistory: []user.LoginHistory{
			{
				LoginTime:  RandomTime,
				LogoutTime: RandomTime,
			},
		},
		CreatedAt: RandomTime,
		UpdatedAt: RandomTime,
	}
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	os.Setenv("DB_DATABASE", "test")
	os.Setenv("DB_USERNAME", "test")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_HOST", "localhost")
	migrations, err := migrations.GetMigrations()
	if err != nil {
		log.Fatalf("Error getting migrations: %v", err)
	}
	migrations = append(migrations, path.Join("migrations", "test", "get_user.sql"))

	postgresContainer, err := database.GetPostgresTestContainer(
		ctx,
		migrations,
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Error getting postgres container: %v", err)
	}
	defer func() {
		err := postgresContainer.Terminate(ctx)
		if err != nil {
			fmt.Printf("Error terminating postgres container: %v", err)
		}
	}()

	port, _ := postgresContainer.MappedPort(ctx, "5432")
	log.Printf("Postgres container running on port: %s", port.Port())
	err = os.Setenv("DB_PORT", port.Port())
	if err != nil {
		log.Fatalf("Error setting POSTGRES_PORT: %v", err)
	}
	m.Run()
}

func TestPostgresRepository_fetchUser(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		id           sql.NullString
		username     sql.NullString
		email        sql.NullString
		authTye      sql.NullInt16
		Role         sql.NullInt16
		loginHistory sql.NullString
		role         sql.NullInt16
		status       sql.NullInt16
		createdAt    sql.NullTime
		updatedAt    sql.NullTime
		deleteAt     sql.NullTime
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    user.User
		wantErr bool
	}{
		{
			name: "Test fetchUser",
			fields: fields{
				db: nil,
			},
			args: args{
				id:       sql.NullString{String: "1", Valid: true},
				username: sql.NullString{String: "John Doe", Valid: true},
				email:    sql.NullString{String: "john@doe", Valid: true},
				loginHistory: sql.NullString{String: fmt.Sprintf(`[{"login_time":"%s","logout_time":"%s"}]`,
					RandomTimeStr, RandomTimeStr),
					Valid: true,
				},
				authTye:   sql.NullInt16{Int16: int16(user.AuthTypeGoogle), Valid: true},
				role:      sql.NullInt16{Int16: int16(auth.RoleAdmin), Valid: true},
				status:    sql.NullInt16{Int16: int16(user.StatusActive), Valid: true},
				createdAt: sql.NullTime{Time: RandomTime, Valid: true},
				updatedAt: sql.NullTime{Time: RandomTime, Valid: true},
				deleteAt:  sql.NullTime{Time: RandomTime, Valid: true},
			},
			want: user.User{
				Id:       "1",
				Username: "John Doe",
				Email:    "john@doe",
				AuthType: user.AuthTypeGoogle,
				LoginHistory: []user.LoginHistory{
					{
						LoginTime:  RandomTime,
						LogoutTime: RandomTime,
					},
				},
				Role:      auth.RoleAdmin,
				Status:    user.StatusActive,
				CreatedAt: RandomTime,
				UpdatedAt: RandomTime,
				DeletedAt: RandomTime,
			},
			wantErr: false,
		},
		{
			name: "Test fetchUser with invalid login history",
			fields: fields{
				db: nil,
			},
			args: args{
				id:           sql.NullString{String: "1", Valid: true},
				username:     sql.NullString{String: "John Doe", Valid: true},
				email:        sql.NullString{String: "john@doe", Valid: true},
				authTye:      sql.NullInt16{Int16: int16(user.AuthTypeGoogle), Valid: true},
				loginHistory: sql.NullString{String: "invalid", Valid: true},
				role:         sql.NullInt16{Int16: 1, Valid: true},
				status:       sql.NullInt16{Int16: 1, Valid: true},
				createdAt:    sql.NullTime{Time: RandomTime, Valid: true},
				updatedAt:    sql.NullTime{Time: RandomTime, Valid: true},
				deleteAt:     sql.NullTime{Time: RandomTime, Valid: true},
			},
			want:    user.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{
				db: tt.fields.db,
			}
			got, err := p.fetchUser(
				tt.args.id,
				tt.args.username,
				tt.args.email,
				tt.args.loginHistory,
				tt.args.role,
				tt.args.status,
				tt.args.authTye,
				tt.args.createdAt,
				tt.args.updatedAt,
				tt.args.deleteAt,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchUser() \nerror = %v\nwantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchUser() \ngot = %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestPostgresRepository_GetUser(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    user.User
		wantErr errs.ChatError
	}{
		{
			name: "Test GetUser with valid key and value",
			args: args{
				ctx:   context.Background(),
				key:   "email",
				value: "john@doe",
			},
			want:    UserJonhDoe,
			wantErr: nil,
		},
		{
			name: "Test GetUser with invalid key",
			args: args{
				ctx:   context.Background(),
				key:   "invalid",
				value: "john@doe",
			},
			wantErr: errs.NewError(
				errs.ErrBadRequest,
				errors.New("invalid key. Valid keys are id, username, email"),
			),
		},
		{
			name: "Test GetUser with invalid value",
			args: args{
				ctx:   context.Background(),
				key:   "email",
				value: "",
			},
			wantErr: errs.NewError(errs.ErrBadRequest, errors.New("invalid value")),
		},
		{
			name: "Test GetUser with not found user",
			args: args{
				ctx:   context.Background(),
				key:   "email",
				value: "notfound@doe",
			},
			wantErr: errs.NewError(
				errs.ErrNotFound,
				fmt.Errorf("user with %s=%s not found", "email", "notfound@doe"),
			),
		},
	}
	db, err := database.GetPostgresConn(false)
	if err != nil {
		log.Fatalf("Error getting postgres connection: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{
				db: db,
			}
			got, err := p.GetUser(tt.args.ctx, tt.args.key, tt.args.value)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("GetUser()\n\terror = %v\n\twantErr = %v", err, tt.wantErr)
				return
			}
			if got.Username != tt.want.Username || got.Email != tt.want.Email || got.AuthType != tt.want.AuthType ||
				got.Role != tt.want.Role || got.Status != tt.want.Status || got.CreatedAt != tt.want.CreatedAt {
				t.Errorf("GetUser() \n\tgot = %v,\n\twant %v", got, tt.want)
			}
		})
	}
}

func TestPostgresRepository_CreateUser(t *testing.T) {
	type args struct {
		u user.User
	}
	tests := []struct {
		name    string
		args    args
		want    user.User
		wantErr errs.ChatError
	}{
		{
			name: "Test CreateUser",
			args: args{
				u: UserTomDoe,
			},
			want:    UserTomDoe,
			wantErr: nil,
		},
		{
			name: "Test CreateUser with user already exists",
			args: args{
				u: UserJonhDoe,
			},
			want: user.User{},
			wantErr: errs.NewError(
				errs.ErrConflict,
				fmt.Errorf("user with %s=%s already exists", "username", UserJonhDoe.Username),
			),
		},
		{
			name: "Test CreateUser with email already exists",
			args: args{
				u: UserJohnnDoe,
			},
			want: user.User{},
			wantErr: errs.NewError(
				errs.ErrConflict,
				fmt.Errorf("user with %s=%s already exists", "email", UserJohnnDoe.Email),
			),
		},
	}
	db, err := database.GetPostgresConn(false)
	if err != nil {
		t.Fatalf("Error getting postgres connection: %v", err)
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{db: db}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatalf("Error beginning transaction: %v", err)
			}
			got, err := p.CreateUser(ctx, tx, tt.args.u)
			tx.Commit()
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("CreateUser() \n\terror = %v\n\twantErr = %v", err, tt.wantErr)
				return
			}
			if got.Username != tt.want.Username || got.Email != tt.want.Email || got.AuthType != tt.want.AuthType {
				t.Errorf("CreateUser() \n\tgot = %v\n\twant = %v", got, tt.want)
			}
		})
	}
}

func TestPostgresRepository_UpdateUser(t *testing.T) {
	type args struct {
		u user.User
	}
	tests := []struct {
		name    string
		args    args
		want    user.User
		wantErr bool
	}{
		{
			name: "Test UpdateUser",
			args: args{
				u: UserJonhDoeUpdated,
			},
		},
	}
	db, err := database.GetPostgresConn(false)
	if err != nil {
		t.Fatalf("Error getting postgres connection: %v", err)
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{
				db: db,
			}
			tx, err := db.BeginTx(ctx, nil)
			got, err := p.UpdateUser(ctx, tx, tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Error committing transaction: %v", err)
			}
			if got.Id != tt.args.u.Id {
				t.Errorf("UpdateUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgresRepository_DeleteUser(t *testing.T) {
	type args struct {
		u user.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test DeleteUser",
			args: args{
				u: UserJonhDoe,
			},
		},
	}
	db, err := database.GetPostgresConn(false)
	if err != nil {
		t.Fatalf("Error getting postgres connection: %v", err)
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{
				db: db,
			}
			tx, err := db.BeginTx(ctx, nil)
			got, err := p.DeleteUser(ctx, tx, tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Error committing transaction: %v", err)
			}
			if got.DeletedAt.IsZero() {
				t.Errorf("DeleteUser() got = %v, want %v", got, tt.args.u)
			}
		})
	}
}

func TestPostgresRepository_ListUsers(t *testing.T) {
	type args struct {
		columns []string
		filters []user.Filter
		sorts   []user.Sort
		page    user.Pagination
	}
	tests := []struct {
		name       string
		args       args
		wantLength int
		wantErr    bool
	}{
		{
			name: "Test ListUsers",
			args: args{
				columns: []string{"id", "username"},
				filters: []user.Filter{},
				sorts:   []user.Sort{},
				page: user.Pagination{
					Offset: 0,
					Limit:  10,
				},
			},
			wantLength: 4,
		},
	}
	db, err := database.GetPostgresConn(false)
	if err != nil {
		t.Fatalf("Error getting postgres connection: %v", err)
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PostgresRepository{
				db: db,
			}
			got, err := p.ListUsers(ctx, tt.args.columns, tt.args.filters, tt.args.sorts, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLength {
				t.Errorf("ListUsers() got = %v, want %v", len(got), tt.wantLength)
			}
		})
	}
}

func Test_buildListQuery(t *testing.T) {
	type args struct {
		columns []string
		filters []user.Filter
		sorts   []user.Sort
		page    user.Pagination
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test buildListQuery",
			args: args{
				columns: []string{"id"},
				filters: []user.Filter{
					{
						Key:        "role",
						Value:      1,
						Comparison: user.ComparisonEqual,
					},
				},
				sorts: nil,

				page: user.Pagination{
					Offset: 0,
					Limit:  10,
				},
			},
			want:    "SELECT id FROM public.user WHERE role = $1 ORDER BY created_at DESC LIMIT 10 OFFSET 0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := buildListQuery(tt.args.columns, tt.args.filters, tt.args.sorts, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildListQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildListQuery() got = %v, want %v", got, tt.want)
			}
		})
	}
}
