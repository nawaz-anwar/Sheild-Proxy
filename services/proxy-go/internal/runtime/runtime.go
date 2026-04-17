package runtime

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Redis    *RedisClient
	Postgres *PostgresClient
}

type RedisClient struct {
	Client *redis.Client
	Addr   string
	Prefix string
}

type PostgresClient struct {
	Pool *pgxpool.Pool
	DSN  string
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
		_ = d.Redis.Client.Close()
		log.Printf("runtime: redis client closed addr=%s", d.Redis.Addr)
	}
	if d.Postgres != nil {
		d.Postgres.Pool.Close()
		log.Printf("runtime: postgres client closed")
	}
}

func initRedis(cfg config.RedisConfig) (*RedisClient, error) {
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	opt, err := redisOptions(addr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	log.Printf("runtime: redis initialized addr=%s", addr)
	return &RedisClient{Client: client, Addr: addr, Prefix: strings.TrimSpace(cfg.KeyPrefix)}, nil
}

func initPostgres(cfg config.PostgresConfig) (*PostgresClient, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}
	log.Printf("runtime: postgres initialized")
	return &PostgresClient{Pool: pool, DSN: dsn}, nil
}

func redisOptions(addr string) (*redis.Options, error) {
	normalized := strings.TrimSpace(addr)
	if normalized == "" {
		return nil, fmt.Errorf("redis addr is required")
	}
	if strings.Contains(normalized, "://") {
		opt, err := redis.ParseURL(normalized)
		if err != nil {
			return nil, fmt.Errorf("invalid redis url: %w", err)
		}
		return opt, nil
	}
	if strings.Contains(normalized, "@") {
		return nil, fmt.Errorf("redis addr with credentials must use redis:// URL format")
	}
	if _, err := url.Parse("redis://" + normalized); err != nil {
		return nil, fmt.Errorf("invalid redis addr: %w", err)
	}
	return &redis.Options{Addr: normalized}, nil
}
