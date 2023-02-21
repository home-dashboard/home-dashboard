package monitor_model

import "encoding/gob"

type UserRole = int

var (
	RoleAdministrator UserRole = 1
	RoleGuest                  = 2
)

type User struct {
	Model
	Username string   `json:"username"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

func init() {
	gob.Register(User{})
}
