package domain

type UserTraffic struct {
	RxBytes uint64
	TxBytes uint64
}

type TrafficSnapshot struct {
	Users  map[string]UserTraffic
	Online map[string]bool
}
