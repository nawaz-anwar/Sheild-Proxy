package runtime

import (
	"fmt"
	"log"
	"strings"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
)

type Dependencies struct {
	Redis    *RedisClient
	Postgres *PostgresClient
}

type RedisClient struct {
	Addr string
}

type PostgresClient struct {
	DSN string
}

func Init(cfg *config.Config) (*Dependencies, error) {
	redis, err := initRedis(cfg.Redis)
	if err != nil {
		return nil, err
	}
	pg, err := initPostgres(cfg.Postgres)
	if err != nil {
		return nil, err
	}
	return &Dependencies{Redis: redis, Postgres: pg}, nil
}

func (d *Dependencies) Close() {
	if d == nil {
		return
	}
	if d.Redis != nil {
		log.Printf("runtime: redis client closed addr=%s", d.Redis.Addr)
	}
	if d.Postgres != nil {
		log.Printf("runtime: postgres client closed")
	}
}

func initRedis(cfg config.RedisConfig) (*RedisClient, error) {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	log.Printf("runtime: redis initialized addr=%s", addr)
	return &RedisClient{Addr: addr}, nil
}

func initPostgres(cfg config.PostgresConfig) (*PostgresClient, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	log.Printf("runtime: postgres initialized")
	return &PostgresClient{DSN: dsn}, nil
}
