package server

import (
	"crypto/subtle"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	wsBearerSubprotocolPrefix = "eth-amm-sim.bearer."

	defaultRateLimitRead     = 60
	defaultRateLimitMutation = 10
	rateLimitWindow          = time.Minute
	rateLimitIdleTTL         = 5 * time.Minute
)

// securityConfig holds runtime auth / rate-limit settings loaded from env.
type securityConfig struct {
	apiToken           string
	trustXForwardedFor bool
	readLimit          int
	mutationLimit      int
}

func loadSecurityConfig() securityConfig {
	cfg := securityConfig{
		apiToken:           strings.TrimSpace(os.Getenv("ETH_AMM_SIM_API_TOKEN")),
		trustXForwardedFor: envTruthy("ETH_AMM_SIM_TRUST_X_FORWARDED_FOR"),
		readLimit:          envIntOr("ETH_AMM_SIM_RATE_LIMIT_READ", defaultRateLimitRead),
		mutationLimit:      envIntOr("ETH_AMM_SIM_RATE_LIMIT_MUTATION", defaultRateLimitMutation),
	}
	if cfg.readLimit < 1 {
		cfg.readLimit = defaultRateLimitRead
	}
	if cfg.mutationLimit < 1 {
		cfg.mutationLimit = defaultRateLimitMutation
	}
	return cfg
}

func envTruthy(key string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

func envIntOr(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func (s *Server) logAuthStartupWarning() {
	if s.security.apiToken == "" {
		log.Printf("WARNING: ETH_AMM_SIM_API_TOKEN is unset - mutation and WebSocket authentication is disabled (local-dev mode)")
	}
}

func (s *Server) authRequired() bool {
	return s.security.apiToken != ""
}

func (s *Server) tokenValid(candidate string) bool {
	if s.security.apiToken == "" || candidate == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(candidate), []byte(s.security.apiToken)) == 1
}

// bearerTokenFromHeader extracts a Bearer token from Authorization.
// Query parameters are never accepted.
func bearerTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(auth[len(prefix):])
}

// wsCredentials returns the presented token and an optional subprotocol to
// negotiate. Prefer Authorization: Bearer; otherwise accept
// Sec-WebSocket-Protocol: eth-amm-sim.bearer.<token>.
func wsCredentials(r *http.Request) (token string, negotiate string) {
	if t := bearerTokenFromHeader(r); t != "" {
		return t, ""
	}
	for _, p := range websocket.Subprotocols(r) {
		if strings.HasPrefix(p, wsBearerSubprotocolPrefix) {
			return strings.TrimPrefix(p, wsBearerSubprotocolPrefix), p
		}
	}
	return "", ""
}

func (s *Server) authorizeRequest(r *http.Request) bool {
	if !s.authRequired() {
		return true
	}
	return s.tokenValid(bearerTokenFromHeader(r))
}

func (s *Server) authorizeWebSocket(r *http.Request) (ok bool, negotiate string) {
	if !s.authRequired() {
		return true, ""
	}
	token, negotiate := wsCredentials(r)
	if !s.tokenValid(token) {
		return false, ""
	}
	return true, negotiate
}

func isPublicHealthPath(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func isMutationPath(method, path string) bool {
	if method == http.MethodOptions {
		return false
	}
	if path == "/stream" {
		return true
	}
	if method != http.MethodPost {
		return false
	}
	switch path {
	case "/session/start", "/session/pause", "/session/resume", "/session/stop", "/session/reset",
		"/trade/buy", "/trade/sell":
		return true
	default:
		return false
	}
}

func requiresAuth(method, path string) bool {
	if method == http.MethodOptions || isPublicHealthPath(path) {
		return false
	}
	if path == "/stream" {
		return true
	}
	return isMutationPath(method, path)
}

// ── Rate limiting ────────────────────────────────────────────────────────────

type rateWindow struct {
	count   int
	resetAt time.Time
}

type ipRateState struct {
	read     rateWindow
	mut      rateWindow
	lastSeen time.Time
}

type rateLimiter struct {
	mu            sync.Mutex
	clients       map[string]*ipRateState
	readLimit     int
	mutationLimit int
	window        time.Duration
	idleTTL       time.Duration
	lastCleanup   time.Time
}

func newRateLimiter(readLimit, mutationLimit int) *rateLimiter {
	return &rateLimiter{
		clients:       make(map[string]*ipRateState),
		readLimit:     readLimit,
		mutationLimit: mutationLimit,
		window:        rateLimitWindow,
		idleTTL:       rateLimitIdleTTL,
	}
}

func (rl *rateLimiter) allow(ip string, mutation bool, now time.Time) (ok bool, retryAfter time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if now.Sub(rl.lastCleanup) > rl.idleTTL {
		rl.cleanupLocked(now)
		rl.lastCleanup = now
	}

	st, okClient := rl.clients[ip]
	if !okClient {
		st = &ipRateState{}
		rl.clients[ip] = st
	}
	st.lastSeen = now

	win := &st.read
	limit := rl.readLimit
	if mutation {
		win = &st.mut
		limit = rl.mutationLimit
	}

	if win.resetAt.IsZero() || !now.Before(win.resetAt) {
		win.count = 0
		win.resetAt = now.Add(rl.window)
	}
	if win.count >= limit {
		ra := win.resetAt.Sub(now)
		if ra < time.Second {
			ra = time.Second
		}
		return false, ra
	}
	win.count++
	return true, 0
}

func (rl *rateLimiter) cleanupLocked(now time.Time) {
	for ip, st := range rl.clients {
		if now.Sub(st.lastSeen) > rl.idleTTL {
			delete(rl.clients, ip)
		}
	}
}

func clientIP(r *http.Request, trustXFF bool) string {
	if trustXFF {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[0]); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		// Health probes stay public and are not rate-limited.
		if isPublicHealthPath(path) {
			next.ServeHTTP(w, r)
			return
		}

		// CORS preflight: allow through (corsMiddleware handles OPTIONS).
		if method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r, s.security.trustXForwardedFor)
		mutation := isMutationPath(method, path)
		if ok, retryAfter := s.limiter.allow(ip, mutation, time.Now()); !ok {
			secs := int(retryAfter.Seconds())
			if secs < 1 {
				secs = 1
			}
			w.Header().Set("Retry-After", strconv.Itoa(secs))
			respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		// WebSocket auth is enforced inside handleWebSocket (needs subprotocol
		// negotiation on the upgrade response). REST mutations use Bearer here.
		if requiresAuth(method, path) && path != "/stream" {
			if !s.authorizeRequest(r) {
				respondError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
