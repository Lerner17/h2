package rotate_password

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type UseCase struct {
	repo      UserRepository
	restarter ServiceRestarter
	passwords PasswordGenerator
}

func NewUseCase(repo UserRepository, restarter ServiceRestarter, passwords PasswordGenerator) *UseCase {
	return &UseCase{repo: repo, restarter: restarter, passwords: passwords}
}

func (u *UseCase) Execute(ctx context.Context, username string) (string, error) {
	password, err := u.passwords.Generate()
	if err != nil {
		return "", err
	}
	user, err := domain.NewUser(username, password)
	if err != nil {
		return "", err
	}
	if err := u.repo.RotatePassword(ctx, user); err != nil {
		return "", err
	}
	if err := u.restarter.Restart(ctx); err != nil {
		return "", err
	}
	return password, nil
}
