package database

import (
	"time"

	"github.com/go-redis/redis"
)

type Metadata struct {
	rdClient     *redis.Client
	metaLifetime time.Duration
}

func NewMetadata(host string, lifetime time.Duration) (*Metadata, error) {
	opt, err := redis.ParseURL(host)
	if err != nil {
		return nil, err
	}
	db := redis.NewClient(opt)
	return &Metadata{
		rdClient:     db,
		metaLifetime: lifetime,
	}, nil
}

func (m *Metadata) Get(ip string) (string, error) {
	return m.rdClient.Get(ip).Result()
}

func (m *Metadata) Set(ip, value string) error {
	return m.rdClient.Set(ip, value, m.metaLifetime).Err()
}
