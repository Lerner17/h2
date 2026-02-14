package remove_user

import "context"

type UserRepository interface {
	RemoveUser(ctx context.Context, username string) error
}

type ServiceRestarter interface {
	Restart(ctx context.Context) error
}
