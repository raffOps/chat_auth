package auth

import (
	"github.com/dgrijalva/jwt-go"
	"strconv"
)

type RoleId uint

const (
	RoleAdmin RoleId = 1
	RoleUser  RoleId = 2
)

var MapRole = map[RoleId]string{
	RoleAdmin: "admin",
	RoleUser:  "user",
}

var MapRoleString = map[string]RoleId{
	"admin": RoleAdmin,
	"user":  RoleUser,
}

type Claims struct {
	RoleId    RoleId `json:"role"`
	SessionId string `json:"session_id"`
	jwt.StandardClaims
}

// MarshalBinary converts RoleId to a byte slice for Redis storage.
func (r RoleId) MarshalBinary() ([]byte, error) {
	return []byte(strconv.Itoa(int(r))), nil
}

// UnmarshalBinary converts a byte slice back to RoleId.
func (r *RoleId) UnmarshalBinary(data []byte) error {
	parsed, err := strconv.Atoi(string(data))
	if err != nil {
		return err
	}
	*r = RoleId(parsed)
	return nil
}
