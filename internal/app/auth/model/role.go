package auth

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
