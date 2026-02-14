//go:build wireinject
// +build wireinject

package list_users

import (
	"github.com/google/wire"
	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/infra/configrepo"
)

func BuildUseCase(cfg appconfig.Config) (*UseCase, error) {
	wire.Build(
		provideConfigPath,
		configrepo.NewRepository,
		wire.Bind(new(UserRepository), new(*configrepo.Repository)),
		NewUseCase,
	)
	return nil, nil
}
