package list_users

import (
	"context"
	"testing"
)

type repoMock struct{}

func (repoMock) ListUsers(context.Context) ([]string, error) {
	return []string{"alice", "bob"}, nil
}

func TestExecute(t *testing.T) {
	uc := NewUseCase(repoMock{})
	users, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("unexpected users: %#v", users)
	}
}
