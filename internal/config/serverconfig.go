package config

const (
	DefaultServerAddress = "localhost:8080"
)

type ServerConfig struct {
	HashKey *string
	Address string
}
