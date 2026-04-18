package signer

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// oauth2Token is the cached result of a successful token exchange.
type oauth2Token struct {
	accessToken string
	expiresAt   time.Time
}

// oauth2Cache stores access tokens keyed by a stable fingerprint of the
// request that produced them. Concurrent cache misses for the same key
// collapse to a single upstream call via singleflight.
//
// A 60-second safety margin is subtracted from expires_in to avoid handing
// out a token that's about to expire mid-flight. When the IdP omits
// expires_in, we default to a 5-minute TTL.
type oauth2Cache struct {
	mu     sync.Mutex
	tokens map[string]*oauth2Token
	group  singleflight.Group
	nowFn  func() time.Time
}

func newOAuth2Cache() *oauth2Cache {
	return &oauth2Cache{
		tokens: make(map[string]*oauth2Token),
		nowFn:  time.Now,
	}
}

// get returns a live cached token, or (nil, false) on miss/expiry.
func (c *oauth2Cache) get(key string) (*oauth2Token, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	t, ok := c.tokens[key]
	if !ok {
		return nil, false
	}
	// now >= expiresAt counts as expired (safer than strict After).
	if !c.nowFn().Before(t.expiresAt) {
		delete(c.tokens, key)
		return nil, false
	}
	return t, true
}

// put stores a token. expiresIn is in seconds; safety margin is applied.
func (c *oauth2Cache) put(key, accessToken string, expiresIn int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ttl := time.Duration(expiresIn) * time.Second
	const safety = 60 * time.Second
	const defaultTTL = 5 * time.Minute
	if ttl <= 0 {
		ttl = defaultTTL
	}
	effective := ttl - safety
	if effective <= 0 {
		// IdP returned a TTL at or below the safety margin. Store with
		// zero effective TTL so the next get() treats it as expired and
		// re-fetches. The alternative (caching for the full TTL) risks
		// handing out a token that expires mid-flight.
		effective = 0
	}
	c.tokens[key] = &oauth2Token{
		accessToken: accessToken,
		expiresAt:   c.nowFn().Add(effective),
	}
}

// invalidate drops a key. Used when the upstream returns an auth failure
// that suggests the cached token is no longer valid.
func (c *oauth2Cache) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tokens, key)
}

// fetchOrFetch returns a cached token if live; otherwise invokes fetch
// exactly once across concurrent callers (via singleflight) and caches the
// result.
func (c *oauth2Cache) fetchOrFetch(
	key string,
	fetch func() (accessToken string, expiresIn int64, err error),
) (string, error) {
	if t, ok := c.get(key); ok {
		return t.accessToken, nil
	}
	v, err, _ := c.group.Do(key, func() (any, error) {
		// Re-check after acquiring singleflight slot in case a sibling
		// call already populated the cache.
		if t, ok := c.get(key); ok {
			return t.accessToken, nil
		}
		tok, expiresIn, err := fetch()
		if err != nil {
			return "", err
		}
		c.put(key, tok, expiresIn)
		return tok, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// cacheKey is a stable fingerprint derived from the parameters that
// determine which access token a caller is eligible to receive. Must be
// stable across invocations for the same logical request, and MUST NOT
// include secret values directly — we hash an identifying fingerprint of
// the refresh_token rather than the token itself, so cache keys are safe
// to log or expose in error messages.
func cacheKey(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0}) // delimiter to prevent boundary collisions
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sortedJoin produces a deterministic representation of scopes.
func sortedJoin(scopes []string, sep string) string {
	cp := append([]string(nil), scopes...)
	sort.Strings(cp)
	return strings.Join(cp, sep)
}

// fingerprint returns a short hash of a secret value for use in a cache key.
// Lets us detect "refresh_token changed" without storing the token in the
// key directly.
func fingerprint(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}
