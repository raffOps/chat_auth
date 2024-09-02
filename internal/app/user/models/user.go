package user

import (
	"time"

	authModels "github.com/raffops/chat_auth/internal/app/auth/model"
)

type LoginHistory struct {
	LoginTime  time.Time `json:"login_time"`
	LogoutTime time.Time `json:"logout_time"`
}

type User struct {
	Id           string            `json:"id,omitempty" validate:"required,uuid4"`
	Username     string            `json:"name,omitempty" validate:"required,min=5,max=100"`
	Email        string            `json:"email,omitempty"`
	AuthType     AuthTypeId        `json:"auth_type,omitempty" validate:"required, oneof=GOOGLE GITHUB"`
	Role         authModels.RoleId `json:"role,omitempty" validate:"required, oneof=ADMIN USER"`
	Status       StatusId          `json:"status,omitempty" validate:"required, oneof=ACTIVE INACTIVE"`
	LoginHistory []LoginHistory    `json:"login_history,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
	UpdatedAt    time.Time         `json:"updated_at,omitempty"`
	DeletedAt    time.Time         `json:"deleted_at,omitempty"`
}

type StatusId uint

const (
	StatusActive   StatusId = 1
	StatusInactive StatusId = 2
)

var MapStatus = map[StatusId]string{
	StatusActive:   "active",
	StatusInactive: "inactive",
}

var MapStatusString = map[string]StatusId{
	"active":   StatusActive,
	"inactive": StatusInactive,
}

type AuthTypeId uint

const (
	AuthTypeGoogle AuthTypeId = 1
	AuthTypeGithub AuthTypeId = 2
)

var MapAuthType = map[AuthTypeId]string{
	AuthTypeGoogle: "google",
	AuthTypeGithub: "github",
}

var MapAuthTypeString = map[string]AuthTypeId{
	"google": AuthTypeGoogle,
	"github": AuthTypeGithub,
}

type Filter struct {
	Key        string `validate:"required, oneof=role status auth_type"`
	Value      any
	Comparison Comparison
}

type Comparison string

var (
	ComparisonEqual              Comparison = "="
	ComparisonGreaterThan        Comparison = ">"
	ComparisonGreaterThanOrEqual Comparison = ">="
	ComparisonLessThan           Comparison = "<"
	ComparisonLessThanOrEqual    Comparison = "<="
)

type Sort struct {
	Key   string
	Order Order
}

type Order string

var (
	OrderAsc  Order = "ASC"
	OrderDesc Order = "DESC"
)

var (
	ValidColumnsToFetch = []string{
		"id",
		"username",
		"email",
		"auth_type",
		"role",
		"status",
		"login_history",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	ValidColumnsToFilter = []string{"role", "status", "auth_type"}
	ValidColumnsToSort   = []string{"created_at", "updated_at", "deleted_at"}
)

type Pagination struct {
	Limit  int `validate:"required,min=1,max=100"`
	Offset int `validate:"required,min=0"`
}
