package nats

import (
	"errors"
	"strings"

	"github.com/nats-io/nats.go"
)

type Config struct {
	Servers []string
	Cluster string
}

func (c Config) Validate() error {
	if len(c.Servers) == 0 {
		return errors.New("at least one nats server is required")
	}
	for _, server := range c.Servers {
		if strings.TrimSpace(server) == "" {
			return errors.New("nats server cannot be empty")
		}
	}
	return nil
}

func New(cfg Config, opts ...nats.Option) (*nats.Conn, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	servers := strings.Join(cfg.Servers, ",")
	if strings.TrimSpace(cfg.Cluster) != "" {
		opts = append(opts, nats.Name(cfg.Cluster))
	}
	return nats.Connect(servers, opts...)
}
