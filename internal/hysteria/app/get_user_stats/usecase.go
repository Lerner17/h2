package get_user_stats

import "context"

type UserStats struct {
	Online     bool
	RxBytes    uint64
	TxBytes    uint64
	TotalBytes uint64
}

type UseCase struct {
	repo TrafficStatsRepository
}

func NewUseCase(repo TrafficStatsRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (u *UseCase) Execute(ctx context.Context, users []string) (map[string]UserStats, error) {
	stats := make(map[string]UserStats, len(users))
	for _, username := range users {
		stats[username] = UserStats{}
	}

	snapshot, err := u.repo.Fetch(ctx)
	if err != nil {
		return stats, nil
	}

	for _, username := range users {
		current := stats[username]
		if traffic, ok := snapshot.Users[username]; ok {
			current.RxBytes = traffic.RxBytes
			current.TxBytes = traffic.TxBytes
			current.TotalBytes = traffic.RxBytes + traffic.TxBytes
		}
		if online, ok := snapshot.Online[username]; ok {
			current.Online = online
		}
		stats[username] = current
	}

	return stats, nil
}
