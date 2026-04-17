package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	issuer     string
	audience   string
	secret     []byte
	cookieName string
	ttl        time.Duration
}

type claims struct {
	Issuer   string `json:"iss"`
	Audience string `json:"aud"`
	Subject  string `json:"sub"`
	IP       string `json:"ip"`
	UA       string `json:"ua"`
	IssuedAt int64  `json:"iat"`
	Expires  int64  `json:"exp"`
}

func New(issuer, audience, secret, cookieName string, ttlSeconds int) *Service {
	if issuer == "" {
		issuer = "shield-proxy"
	}
	if audience == "" {
		audience = "shield-origin"
	}
	if secret == "" {
		secret = "shield-local-dev-secret"
	}
	if cookieName == "" {
		cookieName = "shield_token"
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 900
	}
	return &Service{
		issuer:     issuer,
		audience:   audience,
		secret:     []byte(secret),
		cookieName: cookieName,
		ttl:        time.Duration(ttlSeconds) * time.Second,
	}
}

func (s *Service) CookieName() string {
	return s.cookieName
}

func (s *Service) TTL() time.Duration {
	return s.ttl
}

func (s *Service) Issue(domain, ip, ua string, now time.Time) (string, error) {
	hdr := map[string]string{"alg": "HS256", "typ": "JWT"}
	payload := claims{
		Issuer:   s.issuer,
		Audience: s.audience,
		Subject:  domain,
		IP:       ip,
		UA:       ua,
		IssuedAt: now.Unix(),
		Expires:  now.Add(s.ttl).Unix(),
	}
	headerJSON, err := json.Marshal(hdr)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	sig := signHS256(unsigned, s.secret)
	return unsigned + "." + sig, nil
}

func (s *Service) Validate(token, ip, ua string, now time.Time) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	unsigned := parts[0] + "." + parts[1]
	expected := signHS256(unsigned, s.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return false
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	var c claims
	if err := json.Unmarshal(payloadJSON, &c); err != nil {
		return false
	}
	if c.Issuer != s.issuer || c.Audience != s.audience {
		return false
	}
	if c.IP != ip || c.UA != ua {
		return false
	}
	if now.Unix() >= c.Expires {
		return false
	}
	return true
}

func signHS256(unsigned string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func ParseChallengeHeaders(id, answer string) (string, string, error) {
	id = strings.TrimSpace(id)
	answer = strings.TrimSpace(answer)
	if id == "" || answer == "" {
		return "", "", errors.New("missing challenge id or answer")
	}
	return id, answer, nil
}

func BearerToken(header string) (string, error) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}
