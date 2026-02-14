package get_user_stats

import (
	"time"

	appconfig "vpn/internal/config"
)

func provideTrafficStatsEnabled(cfg appconfig.Config) bool {
	return cfg.HysteriaTrafficStatsEnabled
}

func provideTrafficStatsURL(cfg appconfig.Config) string {
	return cfg.HysteriaTrafficStatsURL
}

func provideTrafficStatsSecret(cfg appconfig.Config) string {
	return cfg.HysteriaTrafficStatsSecret
}

func provideTrafficStatsTimeout(cfg appconfig.Config) time.Duration {
	if cfg.HysteriaTrafficStatsTimeoutSeconds <= 0 {
		return 2 * time.Second
	}
	return time.Duration(cfg.HysteriaTrafficStatsTimeoutSeconds) * time.Second
}
