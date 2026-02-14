package domain

type ConnectionConfig struct {
	Username     string
	Password     string
	Host         string
	Port         int
	SNI          string
	ObfsType     string
	ObfsPassword string
}
