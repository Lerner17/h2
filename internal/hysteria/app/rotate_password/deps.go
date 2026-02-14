package rotate_password

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type UserRepository interface {
	RotatePassword(ctx context.Context, user domain.User) error
}

type ServiceRestarter interface {
	Restart(ctx context.Context) error
}

type PasswordGenerator interface {
	Generate() (string, error)
}
