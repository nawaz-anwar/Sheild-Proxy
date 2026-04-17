package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/config"
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
	DB  *sql.DB
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
		_ = d.Redis.Client.Close()
		log.Printf("runtime: redis client closed addr=%s", d.Redis.Addr)
	}
	if d.Postgres != nil {
		_ = d.Postgres.DB.Close()
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
		return nil, fmt.Errorf("redis ping failed addr=%s: %w", addr, err)
	}
	log.Printf("runtime: redis initialized addr=%s", addr)
	return &RedisClient{Client: client, Addr: addr, Prefix: strings.TrimSpace(cfg.KeyPrefix)}, nil
}

func initPostgres(cfg config.PostgresConfig) (*PostgresClient, error) {
	dsn := strings.TrimSpace(cfg.DSN)
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("create postgres connection pool: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres ping failed dsn=%s: %w", sanitizeDSN(dsn), err)
	}
	log.Printf("runtime: postgres initialized")
	return &PostgresClient{DB: db, DSN: dsn}, nil
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

func sanitizeDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "<invalid-dsn>"
	}
	if u.User != nil {
		username := u.User.Username()
		if username != "" {
			u.User = url.UserPassword(username, "******")
		} else {
			u.User = nil
		}
	}
	return u.String()
}
