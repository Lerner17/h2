package remove_user

import (
	"context"
	"testing"
)

type repoMock struct{ called bool }

func (m *repoMock) RemoveUser(_ context.Context, _ string) error {
	m.called = true
	return nil
}

type restarterMock struct{ called bool }

func (m *restarterMock) Restart(context.Context) error {
	m.called = true
	return nil
}

func TestExecute(t *testing.T) {
	repo := &repoMock{}
	restarter := &restarterMock{}
	uc := NewUseCase(repo, restarter)

	if err := uc.Execute(context.Background(), "alice"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.called || !restarter.called {
		t.Fatal("expected repo and restarter calls")
	}
}
