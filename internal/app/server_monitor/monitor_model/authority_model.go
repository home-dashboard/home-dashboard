package monitor_model

import (
	"encoding/gob"
	"github.com/siaikin/home-dashboard/internal/pkg/authority"
)

type UserRole = int

var (
	RoleAdministrator UserRole = 1
	RoleGuest                  = 2
)

type User struct {
	Model `json:"model"`
	authority.User
	Role UserRole `json:"role"`
}

func init() {
	gob.Register(authority.User{})
	gob.Register(User{})
}
