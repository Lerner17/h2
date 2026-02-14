package add_user

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type UserRepository interface {
	AddUser(ctx context.Context, user domain.User) error
}

type ServiceRestarter interface {
	Restart(ctx context.Context) error
}

type PasswordGenerator interface {
	Generate() (string, error)
}
