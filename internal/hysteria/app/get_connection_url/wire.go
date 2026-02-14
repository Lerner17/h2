//go:build wireinject
// +build wireinject

package get_connection_url

import (
	"github.com/google/wire"
	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/infra/configrepo"
)

func BuildUseCase(cfg appconfig.Config) (*UseCase, error) {
	wire.Build(
		provideConfigPath,
		configrepo.NewRepository,
		wire.Bind(new(ConnectionRepository), new(*configrepo.Repository)),
		NewUseCase,
	)
	return nil, nil
}
