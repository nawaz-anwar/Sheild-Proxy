package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type Signer struct {
	enabled         bool
	secret          []byte
	header          string
	timestampHeader string
}

func New(enabled bool, secret, header, timestampHeader string) *Signer {
	if header == "" {
		header = "X-Shield-Signature"
	}
	if timestampHeader == "" {
		timestampHeader = "X-Shield-Signature-Timestamp"
	}
	return &Signer{
		enabled:         enabled,
		secret:          []byte(secret),
		header:          header,
		timestampHeader: timestampHeader,
	}
}

func (s *Signer) Sign(method, path, host, clientID string, now time.Time) (header, timestampHeader, signature, timestamp string, ok bool) {
	if s == nil || !s.enabled || len(s.secret) == 0 {
		return "", "", "", "", false
	}
	timestamp = now.UTC().Format(time.RFC3339)
	payload := strings.Join([]string{method, path, host, clientID, timestamp}, "\n")

	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	signature = hex.EncodeToString(mac.Sum(nil))
	return s.header, s.timestampHeader, signature, timestamp, true
}
