//go:build wireinject
// +build wireinject

package remove_user

import (
	"github.com/google/wire"
	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/infra/configrepo"
	"vpn/internal/hysteria/infra/servicectl"
)

func BuildUseCase(cfg appconfig.Config) (*UseCase, error) {
	wire.Build(
		provideConfigPath,
		provideServiceName,
		provideRestartEnabled,
		provideRestartCommand,
		configrepo.NewRepository,
		servicectl.NewRestarter,
		wire.Bind(new(UserRepository), new(*configrepo.Repository)),
		wire.Bind(new(ServiceRestarter), new(*servicectl.Restarter)),
		NewUseCase,
	)
	return nil, nil
}
