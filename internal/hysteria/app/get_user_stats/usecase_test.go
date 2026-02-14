package get_user_stats

import (
	"context"
	"errors"
	"testing"

	"vpn/internal/hysteria/domain"
)

type fakeRepo struct {
	snapshot domain.TrafficSnapshot
	err      error
}

func (f fakeRepo) Fetch(context.Context) (domain.TrafficSnapshot, error) {
	if f.err != nil {
		return domain.TrafficSnapshot{}, f.err
	}
	return f.snapshot, nil
}

func TestExecuteReturnsDefaultsOnError(t *testing.T) {
	uc := NewUseCase(fakeRepo{err: errors.New("boom")})

	stats, err := uc.Execute(context.Background(), []string{"alice"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := stats["alice"]; got.Online || got.RxBytes != 0 || got.TxBytes != 0 || got.TotalBytes != 0 {
		t.Fatalf("expected zero stats, got %+v", got)
	}
}

func TestExecuteMapsData(t *testing.T) {
	uc := NewUseCase(fakeRepo{snapshot: domain.TrafficSnapshot{
		Users:  map[string]domain.UserTraffic{"alice": {RxBytes: 10, TxBytes: 5}},
		Online: map[string]bool{"alice": true},
	}})

	stats, err := uc.Execute(context.Background(), []string{"alice", "bob"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got := stats["alice"]; !got.Online || got.RxBytes != 10 || got.TxBytes != 5 || got.TotalBytes != 15 {
		t.Fatalf("unexpected alice stats: %+v", got)
	}
	if got := stats["bob"]; got.Online || got.TotalBytes != 0 {
		t.Fatalf("unexpected bob stats: %+v", got)
	}
}
