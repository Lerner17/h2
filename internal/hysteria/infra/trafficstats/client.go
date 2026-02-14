package trafficstats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"vpn/internal/hysteria/domain"
)

type Client struct {
	enabled bool
	url     string
	secret  string
	http    *http.Client
}

func NewClient(enabled bool, url, secret string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Client{
		enabled: enabled,
		url:     strings.TrimRight(url, "/"),
		secret:  secret,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Fetch(ctx context.Context) (domain.TrafficSnapshot, error) {
	empty := domain.TrafficSnapshot{Users: map[string]domain.UserTraffic{}, Online: map[string]bool{}}
	if !c.enabled || c.url == "" {
		return empty, nil
	}

	trafficRaw, err := c.getJSON(ctx, c.url+"/traffic")
	if err != nil {
		return empty, err
	}
	onlineRaw, err := c.getJSON(ctx, c.url+"/online")
	if err != nil {
		return empty, err
	}

	users := extractTrafficUsers(trafficRaw)
	online := extractOnlineUsers(onlineRaw)
	if users == nil {
		users = map[string]domain.UserTraffic{}
	}
	if online == nil {
		online = map[string]bool{}
	}

	return domain.TrafficSnapshot{Users: users, Online: online}, nil
}

func (c *Client) getJSON(ctx context.Context, url string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.secret != "" {
		req.Header.Set("Authorization", c.secret)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	var payload any
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractTrafficUsers(payload any) map[string]domain.UserTraffic {
	root, ok := payload.(map[string]any)
	if !ok {
		return nil
	}

	candidates := []map[string]any{root}
	for _, key := range []string{"users", "user", "traffic", "perUser", "userTraffic"} {
		if sub, ok := root[key].(map[string]any); ok {
			candidates = append(candidates, sub)
		}
	}

	for _, m := range candidates {
		result := map[string]domain.UserTraffic{}
		for username, raw := range m {
			if username == "tx" || username == "rx" || username == "up" || username == "down" {
				continue
			}
			t, ok := parseTrafficEntry(raw)
			if !ok {
				continue
			}
			result[username] = t
		}
		if len(result) > 0 {
			return result
		}
	}

	return map[string]domain.UserTraffic{}
}

func parseTrafficEntry(raw any) (domain.UserTraffic, bool) {
	m, ok := raw.(map[string]any)
	if !ok {
		return domain.UserTraffic{}, false
	}

	rx, okRx := firstUint(m, "rx", "download", "down", "recv", "receive")
	tx, okTx := firstUint(m, "tx", "upload", "up", "sent", "send")
	if !okRx && !okTx {
		return domain.UserTraffic{}, false
	}
	return domain.UserTraffic{RxBytes: rx, TxBytes: tx}, true
}

func extractOnlineUsers(payload any) map[string]bool {
	switch v := payload.(type) {
	case []any:
		out := map[string]bool{}
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out[s] = true
			}
		}
		return out
	case map[string]any:
		for _, key := range []string{"users", "online", "usernames"} {
			if arr, ok := v[key].([]any); ok {
				out := map[string]bool{}
				for _, item := range arr {
					if s, ok := item.(string); ok && s != "" {
						out[s] = true
					}
				}
				return out
			}
		}
		out := map[string]bool{}
		for username, raw := range v {
			if b, ok := raw.(bool); ok {
				out[username] = b
				continue
			}
			if _, ok := raw.(map[string]any); ok {
				out[username] = true
			}
		}
		return out
	default:
		return map[string]bool{}
	}
}

func firstUint(m map[string]any, keys ...string) (uint64, bool) {
	for _, key := range keys {
		if raw, ok := m[key]; ok {
			if n, ok := toUint64(raw); ok {
				return n, true
			}
		}
	}
	return 0, false
}

func toUint64(v any) (uint64, bool) {
	switch n := v.(type) {
	case json.Number:
		i, err := n.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case int64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case uint64:
		return n, true
	case string:
		i, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}
