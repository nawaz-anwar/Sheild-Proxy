package proxy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/domainstore"
	"github.com/nawaz-anwar/Sheild-Proxy/services/proxy-go/internal/runtime"
	"github.com/redis/go-redis/v9"
)

const (
	defaultDomainCacheTTL = 15 * time.Minute
)

type redisDomainProvider struct {
	redis *runtime.RedisClient
	ttl   time.Duration
}

func newRedisDomainProvider(redisClient *runtime.RedisClient) *redisDomainProvider {
	if redisClient == nil || redisClient.Client == nil {
		return nil
	}
	return &redisDomainProvider{redis: redisClient, ttl: defaultDomainCacheTTL}
}

func (p *redisDomainProvider) Get(ctx context.Context, host string) (domainstore.Domain, bool, error) {
	raw, err := p.redis.Client.Get(ctx, p.key(host)).Result()
	if err != nil {
		if err == redis.Nil {
			return domainstore.Domain{}, false, nil
		}
		return domainstore.Domain{}, false, err
	}
	var d domainstore.Domain
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return domainstore.Domain{}, false, err
	}
	d.Host = strings.ToLower(strings.TrimSpace(host))
	return d, true, nil
}

func (p *redisDomainProvider) Set(ctx context.Context, d domainstore.Domain) {
	if p == nil || p.redis == nil || p.redis.Client == nil {
		return
	}
	if d.Host == "" {
		return
	}
	b, err := json.Marshal(d)
	if err != nil {
		return
	}
	_ = p.redis.Client.Set(ctx, p.key(d.Host), b, p.ttl).Err()
}

func (p *redisDomainProvider) Delete(ctx context.Context, host string) {
	if p == nil || p.redis == nil || p.redis.Client == nil {
		return
	}
	_ = p.redis.Client.Del(ctx, p.key(host)).Err()
}

func (p *redisDomainProvider) key(host string) string {
	prefix := strings.TrimSpace(p.redis.Prefix)
	if prefix == "" {
		prefix = "shield"
	}
	return fmt.Sprintf("%s:domain:%s", prefix, strings.ToLower(strings.TrimSpace(host)))
}

type postgresDomainProvider struct {
	pg *runtime.PostgresClient
}

type pgDomainRow struct {
	Host     string
	Upstream string
	ClientID string
	Active   bool
}

func newPostgresDomainProvider(pgClient *runtime.PostgresClient) *postgresDomainProvider {
	if pgClient == nil || pgClient.DB == nil {
		return nil
	}
	return &postgresDomainProvider{pg: pgClient}
}

func (p *postgresDomainProvider) Get(ctx context.Context, host string) (domainstore.Domain, bool, error) {
	row := p.pg.DB.QueryRowContext(
		ctx,
		`SELECT domain, upstream_url, client_id::text, status = 'active'
		   FROM domains
		  WHERE lower(domain) = lower($1)
		    AND verified = TRUE
		    AND proxy_connected = TRUE
		    AND status = 'active'
		  LIMIT 1`,
		strings.ToLower(strings.TrimSpace(host)),
	)

	var rec pgDomainRow
	if err := row.Scan(&rec.Host, &rec.Upstream, &rec.ClientID, &rec.Active); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return domainstore.Domain{}, false, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domainstore.Domain{}, false, nil
		}
		return domainstore.Domain{}, false, err
	}
	return domainstore.Domain{
		Host:     strings.ToLower(strings.TrimSpace(rec.Host)),
		Upstream: strings.TrimSpace(rec.Upstream),
		ClientID: strings.TrimSpace(rec.ClientID),
		Active:   rec.Active,
	}, true, nil
}

func (p *postgresDomainProvider) LoadActive(ctx context.Context) ([]domainstore.Domain, error) {
	rows, err := p.pg.DB.QueryContext(
		ctx,
		`SELECT domain, upstream_url, client_id::text, status = 'active'
		   FROM domains
		  WHERE verified = TRUE
		    AND proxy_connected = TRUE
		    AND status = 'active'`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domainstore.Domain, 0)
	for rows.Next() {
		var rec pgDomainRow
		if err := rows.Scan(&rec.Host, &rec.Upstream, &rec.ClientID, &rec.Active); err != nil {
			return nil, err
		}
		out = append(out, domainstore.Domain{
			Host:     strings.ToLower(strings.TrimSpace(rec.Host)),
			Upstream: strings.TrimSpace(rec.Upstream),
			ClientID: strings.TrimSpace(rec.ClientID),
			Active:   rec.Active,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
