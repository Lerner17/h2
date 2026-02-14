package get_connection_url

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type ConnectionRepository interface {
	GetConnectionConfig(ctx context.Context, username string) (domain.ConnectionConfig, error)
}
