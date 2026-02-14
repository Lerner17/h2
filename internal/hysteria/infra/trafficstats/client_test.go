package trafficstats

import (
	"encoding/json"
	"testing"
)

func TestExtractTrafficUsers(t *testing.T) {
	payload := map[string]any{
		"users": map[string]any{
			"alice": map[string]any{"rx": json.Number("120"), "tx": json.Number("80")},
		},
	}

	users := extractTrafficUsers(payload)
	if users["alice"].RxBytes != 120 || users["alice"].TxBytes != 80 {
		t.Fatalf("unexpected traffic parse: %+v", users["alice"])
	}
}

func TestExtractOnlineUsers(t *testing.T) {
	payload := map[string]any{"users": []any{"alice", "bob"}}
	online := extractOnlineUsers(payload)
	if !online["alice"] || !online["bob"] {
		t.Fatalf("unexpected online parse: %+v", online)
	}
}

func TestExtractOnlineUsersFromCountMap(t *testing.T) {
	payload := map[string]any{
		"alice": json.Number("1"),
		"bob":   json.Number("0"),
	}

	online := extractOnlineUsers(payload)
	if !online["alice"] {
		t.Fatalf("alice should be online: %+v", online)
	}
	if online["bob"] {
		t.Fatalf("bob should be offline: %+v", online)
	}
}
