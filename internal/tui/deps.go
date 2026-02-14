package tui

import (
	"fmt"

	appconfig "vpn/internal/config"
	"vpn/internal/hysteria/app/add_user"
	"vpn/internal/hysteria/app/get_connection_url"
	"vpn/internal/hysteria/app/get_user_stats"
	"vpn/internal/hysteria/app/list_users"
	"vpn/internal/hysteria/app/remove_user"
	"vpn/internal/hysteria/app/rotate_password"
)

type Dependencies struct {
	AddUser        *add_user.UseCase
	RotatePassword *rotate_password.UseCase
	RemoveUser     *remove_user.UseCase
	ListUsers      *list_users.UseCase
	UserStats      *get_user_stats.UseCase
	Connection     *get_connection_url.UseCase
}

func BuildDependencies(cfg appconfig.Config) (*Dependencies, error) {
	addUC, err := add_user.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build add-user usecase: %w", err)
	}

	rotateUC, err := rotate_password.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build rotate-password usecase: %w", err)
	}

	removeUC, err := remove_user.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build remove-user usecase: %w", err)
	}

	listUC, err := list_users.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build list-users usecase: %w", err)
	}

	connectionUC, err := get_connection_url.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build connection usecase: %w", err)
	}
	userStatsUC, err := get_user_stats.BuildUseCase(cfg)
	if err != nil {
		return nil, fmt.Errorf("build user-stats usecase: %w", err)
	}

	return &Dependencies{
		AddUser:        addUC,
		RotatePassword: rotateUC,
		RemoveUser:     removeUC,
		ListUsers:      listUC,
		UserStats:      userStatsUC,
		Connection:     connectionUC,
	}, nil
}
