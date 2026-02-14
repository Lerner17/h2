package get_user_stats

import (
	"context"

	"vpn/internal/hysteria/domain"
)

type TrafficStatsRepository interface {
	Fetch(ctx context.Context) (domain.TrafficSnapshot, error)
}
