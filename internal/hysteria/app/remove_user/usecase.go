package remove_user

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type UseCase struct {
	repo      UserRepository
	restarter ServiceRestarter
}

func NewUseCase(repo UserRepository, restarter ServiceRestarter) *UseCase {
	return &UseCase{repo: repo, restarter: restarter}
}

func (u *UseCase) Execute(ctx context.Context, username string) error {
	if username == "" {
		return domain.ErrEmptyUsername
	}
	if err := u.repo.RemoveUser(ctx, username); err != nil {
		return err
	}
	return u.restarter.Restart(ctx)
}
