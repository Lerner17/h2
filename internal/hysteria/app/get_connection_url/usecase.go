package get_connection_url

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

type UseCase struct {
	repo ConnectionRepository
}

func NewUseCase(repo ConnectionRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) Execute(ctx context.Context, username string) (string, error) {
	cfg, err := u.repo.GetConnectionConfig(ctx, username)
	if err != nil {
		return "", err
	}

	query := make(url.Values)
	if cfg.SNI != "" {
		query.Set("sni", cfg.SNI)
	}
	if cfg.ObfsType != "" {
		query.Set("obfs", cfg.ObfsType)
	}
	if cfg.ObfsPassword != "" {
		query.Set("obfs-password", cfg.ObfsPassword)
	}

	uURL := &url.URL{
		Scheme:   "hy2",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     "/",
		RawQuery: query.Encode(),
	}
	if uURL.Host == "" {
		return "", fmt.Errorf("invalid connection host")
	}
	return uURL.String(), nil
}
