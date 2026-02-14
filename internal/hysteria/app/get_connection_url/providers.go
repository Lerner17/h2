package get_connection_url

import appconfig "vpn/internal/config"

func provideConfigPath(cfg appconfig.Config) string { return cfg.HysteriaConfigPath }
