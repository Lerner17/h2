//go:build wireinject
// +build wireinject

package get_user_stats

import (
	"github.com/google/wire"
	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/infra/trafficstats"
)

func BuildUseCase(cfg appconfig.Config) (*UseCase, error) {
	wire.Build(
		provideTrafficStatsEnabled,
		provideTrafficStatsURL,
		provideTrafficStatsSecret,
		provideTrafficStatsTimeout,
		trafficstats.NewClient,
		wire.Bind(new(TrafficStatsRepository), new(*trafficstats.Client)),
		NewUseCase,
	)
	return nil, nil
}
