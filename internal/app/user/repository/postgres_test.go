package user

import (
	"database/sql"
	"fmt"
	auth "github.com/raffops/auth/internal/app/auth/model"
	userModels "github.com/raffops/auth/internal/app/user/models"
	"reflect"
	"testing"
	"time"
)

var (
	RandomTimeStr = "2024-05-07T12:27:58Z"
	RandomTime    = time.Date(2024, time.May, 7, 12, 27, 58, 0, time.UTC)
)

func TestPostgresRepository_parseUser(t *testing.T) {
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
		args    args
		want    userModels.User
		wantErr bool
	}{
		{
			name: "Test parseUser",
			args: args{
				id:       sql.NullString{String: "1", Valid: true},
				username: sql.NullString{String: "John Doe", Valid: true},
				email:    sql.NullString{String: "john@doe", Valid: true},
				loginHistory: sql.NullString{String: fmt.Sprintf(`[{"login_time":"%s","logout_time":"%s"}]`,
					RandomTimeStr, RandomTimeStr),
					Valid: true,
				},
				authTye:   sql.NullInt16{Int16: int16(userModels.AuthTypeGoogle), Valid: true},
				role:      sql.NullInt16{Int16: int16(auth.RoleAdmin), Valid: true},
				status:    sql.NullInt16{Int16: int16(userModels.StatusActive), Valid: true},
				createdAt: sql.NullTime{Time: RandomTime, Valid: true},
				updatedAt: sql.NullTime{Time: RandomTime, Valid: true},
				deleteAt:  sql.NullTime{Time: RandomTime, Valid: true},
			},
			want: userModels.User{
				Id:       "1",
				Username: "John Doe",
				Email:    "john@doe",
				AuthType: userModels.AuthTypeGoogle,
				LoginHistory: []userModels.LoginHistory{
					{
						LoginTime:  RandomTime,
						LogoutTime: RandomTime,
					},
				},
				Role:      auth.RoleAdmin,
				Status:    userModels.StatusActive,
				CreatedAt: RandomTime,
				UpdatedAt: RandomTime,
				DeletedAt: RandomTime,
			},
			wantErr: false,
		},
		{
			name: "Test parseUser with invalid login history",
			args: args{
				id:           sql.NullString{String: "1", Valid: true},
				username:     sql.NullString{String: "John Doe", Valid: true},
				email:        sql.NullString{String: "john@doe", Valid: true},
				authTye:      sql.NullInt16{Int16: int16(userModels.AuthTypeGoogle), Valid: true},
				loginHistory: sql.NullString{String: "invalid", Valid: true},
				role:         sql.NullInt16{Int16: 1, Valid: true},
				status:       sql.NullInt16{Int16: 1, Valid: true},
				createdAt:    sql.NullTime{Time: RandomTime, Valid: true},
				updatedAt:    sql.NullTime{Time: RandomTime, Valid: true},
				deleteAt:     sql.NullTime{Time: RandomTime, Valid: true},
			},
			want:    userModels.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUser(
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
				t.Errorf("parseUser() \nerror = %v\nwantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseUser() \ngot = %v\nwant %v", got, tt.want)
			}
		})
	}
}
