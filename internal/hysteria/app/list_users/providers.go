package list_users

import appconfig "vpn/internal/config"

func provideConfigPath(cfg appconfig.Config) string { return cfg.HysteriaConfigPath }
