package domainvalidator

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

type DomainStatus struct {
	Domain         string
	Verified       bool
	ProxyConnected bool
	Status         string
	UpstreamURL    string
}

type Validator struct {
	db          *sql.DB
	cache       map[string]*DomainStatus
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
	lastRefresh time.Time
}

func NewValidator(db *sql.DB, cacheTTL time.Duration) *Validator {
	v := &Validator{
		db:       db,
		cache:    make(map[string]*DomainStatus),
		cacheTTL: cacheTTL,
	}
	
	// Initial load
	v.RefreshCache()
	
	// Start background refresh
	go v.backgroundRefresh()
	
	return v
}

func (v *Validator) ValidateDomain(domain string) (*DomainStatus, error) {
	v.cacheMutex.RLock()
	status, exists := v.cache[domain]
	v.cacheMutex.RUnlock()

	if exists {
		return status, nil
	}

	// Cache miss - fetch from database
	return v.fetchDomainStatus(domain)
}

func (v *Validator) IsActive(domain string) bool {
	status, err := v.ValidateDomain(domain)
	if err != nil {
		log.Printf("Domain validation error for %s: %v", domain, err)
		return false
	}

	return status.Verified && status.ProxyConnected && status.Status == "active"
}

func (v *Validator) GetUpstreamURL(domain string) (string, error) {
	status, err := v.ValidateDomain(domain)
	if err != nil {
		return "", err
	}

	if !status.Verified {
		return "", fmt.Errorf("domain not verified: %s", domain)
	}

	if !status.ProxyConnected {
		return "", fmt.Errorf("domain not connected to proxy: %s", domain)
	}

	return status.UpstreamURL, nil
}

func (v *Validator) fetchDomainStatus(domain string) (*DomainStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT domain, verified, proxy_connected, status, upstream_url
		FROM domains
		WHERE domain = $1
	`

	var status DomainStatus
	err := v.db.QueryRowContext(ctx, query, domain).Scan(
		&status.Domain,
		&status.Verified,
		&status.ProxyConnected,
		&status.Status,
		&status.UpstreamURL,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found: %s", domain)
	}

	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Update cache
	v.cacheMutex.Lock()
	v.cache[domain] = &status
	v.cacheMutex.Unlock()

	return &status, nil
}

func (v *Validator) RefreshCache() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT domain, verified, proxy_connected, status, upstream_url
		FROM domains
		WHERE status IN ('verified', 'active')
	`

	rows, err := v.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to refresh cache: %w", err)
	}
	defer rows.Close()

	newCache := make(map[string]*DomainStatus)

	for rows.Next() {
		var status DomainStatus
		err := rows.Scan(
			&status.Domain,
			&status.Verified,
			&status.ProxyConnected,
			&status.Status,
			&status.UpstreamURL,
		)
		if err != nil {
			log.Printf("Error scanning domain row: %v", err)
			continue
		}

		newCache[status.Domain] = &status
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	v.cacheMutex.Lock()
	v.cache = newCache
	v.lastRefresh = time.Now()
	v.cacheMutex.Unlock()

	log.Printf("Domain cache refreshed: %d domains loaded", len(newCache))
	return nil
}

func (v *Validator) backgroundRefresh() {
	ticker := time.NewTicker(v.cacheTTL)
	defer ticker.Stop()

	for range ticker.C {
		if err := v.RefreshCache(); err != nil {
			log.Printf("Background cache refresh failed: %v", err)
		}
	}
}

func (v *Validator) GetCacheStats() map[string]interface{} {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_domains": len(v.cache),
		"last_refresh":   v.lastRefresh,
		"cache_ttl":      v.cacheTTL.String(),
	}
}
