package add_user

import (
	"context"
	"testing"

	"vpn/internal/hysteria/domain"
)

type repoMock struct {
	called bool
	user   domain.User
}

func (m *repoMock) AddUser(_ context.Context, user domain.User) error {
	m.called = true
	m.user = user
	return nil
}

type restarterMock struct{ called bool }

func (m *restarterMock) Restart(context.Context) error {
	m.called = true
	return nil
}

type passwordGeneratorMock struct{}

func (passwordGeneratorMock) Generate() (string, error) {
	return "Abc123Abc123Abc123Abc123Abc123Ab", nil
}

func TestExecute(t *testing.T) {
	repo := &repoMock{}
	restarter := &restarterMock{}
	uc := NewUseCase(repo, restarter, passwordGeneratorMock{})

	password, err := uc.Execute(context.Background(), "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password == "" {
		t.Fatal("expected generated password")
	}
	if !repo.called || !restarter.called {
		t.Fatal("expected repo and restarter calls")
	}
	if repo.user.Username != "alice" || repo.user.Password == "" {
		t.Fatalf("unexpected repo payload: %+v", repo.user)
	}
}
