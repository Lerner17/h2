package get_connection_url

import (
	"context"
	"testing"

	"vpn/internal/hysteria/domain"
)

type repoMock struct{}

func (repoMock) GetConnectionConfig(context.Context, string) (domain.ConnectionConfig, error) {
	return domain.ConnectionConfig{
		Username:     "valera",
		Password:     "321",
		Host:         "v1.fr.lerner.dev",
		Port:         443,
		SNI:          "v1.fr.lerner.dev",
		ObfsType:     "salamander",
		ObfsPassword: "32111",
	}, nil
}

func TestExecute(t *testing.T) {
	uc := NewUseCase(repoMock{})
	got, err := uc.Execute(context.Background(), "valera")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "hy2://valera:321@v1.fr.lerner.dev:443/?obfs=salamander&obfs-password=32111&sni=v1.fr.lerner.dev"
	if got != want {
		t.Fatalf("unexpected url: %s", got)
	}
}
