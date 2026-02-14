package list_users

import "context"

type UserRepository interface {
	ListUsers(ctx context.Context) ([]string, error)
}
