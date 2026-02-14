package tui

import "fmt"

func (m model) contentWidth() int {
	if m.width <= 0 {
		return 80
	}
	return max(40, m.width-2)
}

func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m model) aggregateStats() (online int, rx uint64, tx uint64) {
	for _, username := range m.users {
		stat := m.userStats[username]
		if stat.Online {
			online++
		}
		rx += stat.RxBytes
		tx += stat.TxBytes
	}
	return online, rx, tx
}

func formatBytes(v uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case v >= gb:
		return fmt.Sprintf("%.1fG", float64(v)/float64(gb))
	case v >= mb:
		return fmt.Sprintf("%.1fM", float64(v)/float64(mb))
	case v >= kb:
		return fmt.Sprintf("%.1fK", float64(v)/float64(kb))
	default:
		return fmt.Sprintf("%dB", v)
	}
}
