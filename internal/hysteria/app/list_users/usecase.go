package list_users

import "context"

type UseCase struct {
	repo UserRepository
}

func NewUseCase(repo UserRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) Execute(ctx context.Context) ([]string, error) {
	return u.repo.ListUsers(ctx)
}
