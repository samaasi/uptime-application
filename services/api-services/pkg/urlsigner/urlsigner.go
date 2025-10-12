package urlsigner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var urlSigner *Signer

// Signer is used to sign and validate URLs.
type Signer struct {
	Secret         []byte
	ExpiresParam   string
	SignatureParam string
	ClockSkewGrace time.Duration
}

// New creates a new Signer with a secret key and optional configurations.
func New(secret string, options ...Option) *Signer {
	s := &Signer{
		Secret:         []byte(secret),
		ExpiresParam:   "expires",
		SignatureParam: "signature",
		ClockSkewGrace: 10 * time.Second,
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

// Option is a functional option for configuring Signer.
type Option func(*Signer)

// WithExpiresParam sets a custom name for the expiration parameter.
func WithExpiresParam(name string) Option {
	return func(s *Signer) { s.ExpiresParam = name }
}

// WithSignatureParam sets a custom name for the signature parameter.
func WithSignatureParam(name string) Option {
	return func(s *Signer) { s.SignatureParam = name }
}

// WithClockSkewGrace sets a grace period for expiration checks to handle clock skew.
func WithClockSkewGrace(d time.Duration) Option {
	return func(s *Signer) { s.ClockSkewGrace = d }
}

// Generate creates a signed URL with a given lifetime.
// The originalURL should be a relative path (e.g., "/path?param=val"); absolute URLs are rejected.
func (s *Signer) Generate(originalURL string, lifetime time.Duration) (string, error) {
	u, err := url.Parse(originalURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "" || u.Host != "" {
		return "", fmt.Errorf("originalURL must be relative (no scheme or host): %s", originalURL)
	}

	expires := time.Now().Add(lifetime).Unix()
	query := u.Query()
	query.Set(s.ExpiresParam, strconv.FormatInt(expires, 10))

	u.RawQuery = sortQuery(query)

	mac := hmac.New(sha256.New, s.Secret)
	mac.Write([]byte(u.String()))
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	query.Set(s.SignatureParam, signature)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

// Validate checks if a signed URL is valid.
// Returns true if the signature matches and the URL hasn't expired (with grace period).
func (s *Signer) Validate(signedURL string) (bool, error) {
	u, err := url.Parse(signedURL)
	if err != nil {
		return false, err
	}

	query := u.Query()

	signature := query.Get(s.SignatureParam)
	expiresStr := query.Get(s.ExpiresParam)

	if signature == "" || expiresStr == "" {
		return false, fmt.Errorf("missing signature or expiration")
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid expiration format: %w", err)
	}
	if time.Now().Add(s.ClockSkewGrace).Unix() > expires {
		return false, nil
	}

	query.Del(s.SignatureParam)
	u.RawQuery = sortQuery(query)

	mac := hmac.New(sha256.New, s.Secret)
	mac.Write([]byte(u.String()))
	expectedSignature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}

// sortQuery sorts query parameters alphabetically by key, and values per key.
// This ensures a consistent base string for signing.
func sortQuery(q url.Values) string {
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte('&')
		}

		values := q[k]
		sort.Strings(values)
		for j, v := range values {
			if j > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(url.QueryEscape(k))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}
