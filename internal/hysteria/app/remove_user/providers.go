package remove_user

import appconfig "vpn/internal/config"

func provideConfigPath(cfg appconfig.Config) string     { return cfg.HysteriaConfigPath }
func provideServiceName(cfg appconfig.Config) string    { return cfg.HysteriaServiceName }
func provideRestartEnabled(cfg appconfig.Config) bool   { return cfg.HysteriaRestartEnabled }
func provideRestartCommand(cfg appconfig.Config) string { return cfg.HysteriaRestartCommand }
