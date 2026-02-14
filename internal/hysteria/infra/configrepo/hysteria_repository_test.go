package configrepo

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"vpn/internal/hysteria/domain"
)

func TestRepository_AddUser(t *testing.T) {
	t.Parallel()

	t.Run("adds user and keeps other config sections", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `acme:
  domains:
    - v1.fr.lerner.dev
  email: lerner1796@gmail.com
auth:
  type: "userpass"
  userpass:
    lerner: "123"
masquerade:
  type: proxy
`

		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}

		repo := NewRepository(path)
		err := repo.AddUser(context.Background(), domain.User{Username: "valera", Password: "456"})
		if err != nil {
			t.Fatalf("add user: %v", err)
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read config: %v", err)
		}

		content := string(raw)
		if !strings.Contains(content, "valera: \"456\"") {
			t.Fatalf("new user not found in config: %s", content)
		}

		if !strings.Contains(content, "acme:") || !strings.Contains(content, "masquerade:") {
			t.Fatalf("expected unrelated sections to be preserved: %s", content)
		}
	})

	t.Run("returns conflict error for existing user", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `auth:
  type: "userpass"
  userpass:
    lerner: "123"
`

		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}

		repo := NewRepository(path)
		err := repo.AddUser(context.Background(), domain.User{Username: "lerner", Password: "456"})
		if !errors.Is(err, domain.ErrUserAlreadyExists) {
			t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
		}
	})
}

func TestRepository_GetConnectionConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	seed := `listen: :443
acme:
  domains:
    - v1.fr.lerner.dev
auth:
  type: "userpass"
  userpass:
    valera: "321"
obfs:
  type: salamander
  salamander: "32111"
`

	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	repo := NewRepository(path)
	got, err := repo.GetConnectionConfig(context.Background(), "valera")
	if err != nil {
		t.Fatalf("get connection config: %v", err)
	}

	if got.Username != "valera" || got.Password != "321" {
		t.Fatalf("unexpected credentials: %+v", got)
	}
	if got.Host != "v1.fr.lerner.dev" || got.Port != 443 {
		t.Fatalf("unexpected endpoint: %+v", got)
	}
	if got.SNI != "v1.fr.lerner.dev" {
		t.Fatalf("unexpected sni: %+v", got)
	}
	if got.ObfsType != "salamander" || got.ObfsPassword != "32111" {
		t.Fatalf("unexpected obfs: %+v", got)
	}
}

func TestRepository_RotatePassword(t *testing.T) {
	t.Parallel()

	t.Run("updates existing user password", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `auth:
  type: "userpass"
  userpass:
    tester: "old"
masquerade:
  type: proxy
`
		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}

		repo := NewRepository(path)
		if err := repo.RotatePassword(context.Background(), domain.User{Username: "tester", Password: "new"}); err != nil {
			t.Fatalf("rotate password: %v", err)
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		content := string(raw)
		if !strings.Contains(content, "tester: \"new\"") {
			t.Fatalf("updated password not found in config: %s", content)
		}
		if !strings.Contains(content, "masquerade:") {
			t.Fatalf("expected unrelated sections preserved: %s", content)
		}
	})

	t.Run("returns not found for missing user", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `auth:
  type: "userpass"
  userpass:
    tester: "old"
`
		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}

		repo := NewRepository(path)
		err := repo.RotatePassword(context.Background(), domain.User{Username: "ghost", Password: "new"})
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})
}

func TestRepository_RemoveUser(t *testing.T) {
	t.Parallel()

	t.Run("removes existing user", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `auth:
  type: "userpass"
  userpass:
    alice: "111"
    bob: "222"
`
		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}
		repo := NewRepository(path)
		if err := repo.RemoveUser(context.Background(), "bob"); err != nil {
			t.Fatalf("remove user: %v", err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read config: %v", err)
		}
		content := string(raw)
		if strings.Contains(content, "bob: \"222\"") {
			t.Fatalf("user should be removed: %s", content)
		}
	})

	t.Run("returns not found for missing user", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		seed := `auth:
  type: "userpass"
  userpass:
    alice: "111"
`
		if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
			t.Fatalf("write seed: %v", err)
		}
		repo := NewRepository(path)
		err := repo.RemoveUser(context.Background(), "ghost")
		if !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got: %v", err)
		}
	})
}

func TestRepository_ListUsers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	seed := `auth:
  type: "userpass"
  userpass:
    valera: "111"
    alice: "222"
`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	repo := NewRepository(path)
	users, err := repo.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 2 || users[0] != "alice" || users[1] != "valera" {
		t.Fatalf("unexpected users: %#v", users)
	}
}
