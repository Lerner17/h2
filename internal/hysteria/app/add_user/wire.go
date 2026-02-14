//go:build wireinject
// +build wireinject

package add_user

import (
	"github.com/google/wire"
	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/infra/configrepo"
	"vpn/internal/hysteria/infra/servicectl"
	utilpasswordgen "vpn/internal/utils/passwordgen"
)

func BuildUseCase(cfg appconfig.Config) (*UseCase, error) {
	wire.Build(
		provideConfigPath,
		provideServiceName,
		provideRestartEnabled,
		provideRestartCommand,
		configrepo.NewRepository,
		servicectl.NewRestarter,
		utilpasswordgen.NewGenerator,
		wire.Bind(new(UserRepository), new(*configrepo.Repository)),
		wire.Bind(new(ServiceRestarter), new(*servicectl.Restarter)),
		wire.Bind(new(PasswordGenerator), new(*utilpasswordgen.Generator)),
		NewUseCase,
	)
	return nil, nil
}
