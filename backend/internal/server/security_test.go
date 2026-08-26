// security_test.go - Auth, CORS, rate-limit, and WebSocket subprotocol checks.
package server_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"eth-amm-sim/internal/engine"
	"eth-amm-sim/internal/server"
	"eth-amm-sim/internal/store"

	"github.com/gorilla/websocket"
)

const testAPIToken = "test-secret-token-do-not-leak"

func handlerWithEnv(t *testing.T, env map[string]string) http.Handler {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
	// Clear keys not in env so leftover process env cannot leak into tests.
	for _, k := range []string{
		"ETH_AMM_SIM_API_TOKEN",
		"ETH_AMM_SIM_ALLOWED_ORIGINS",
		"ETH_AMM_SIM_TRUST_X_FORWARDED_FOR",
		"ETH_AMM_SIM_RATE_LIMIT_READ",
		"ETH_AMM_SIM_RATE_LIMIT_MUTATION",
	} {
		if _, ok := env[k]; !ok {
			t.Setenv(k, "")
		}
	}
	orch := engine.NewOrchestrator()
	sess := engine.NewSession(orch)
	mem := store.NewMemoryStore()
	srv := server.NewServer(sess, mem, nil)
	return srv.Handler()
}

func postPause(h http.Handler, authHeader string, remoteAddr string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "/session/pause", nil)
	if authHeader != "" {
		r.Header.Set("Authorization", authHeader)
	}
	if remoteAddr != "" {
		r.RemoteAddr = remoteAddr
	} else {
		r.RemoteAddr = "127.0.0.1:40000"
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func getSessionState(h http.Handler, origin, remoteAddr string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/session/state", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if remoteAddr != "" {
		r.RemoteAddr = remoteAddr
	} else {
		r.RemoteAddr = "127.0.0.1:40001"
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// ── API token ────────────────────────────────────────────────────────────────

func TestAPIToken_missingReturns401(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	w := postPause(h, "", "10.0.0.1:1111")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), testAPIToken) {
		t.Fatal("token must not appear in response body")
	}
}

func TestAPIToken_invalidReturns401(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	w := postPause(h, "Bearer wrong-token", "10.0.0.2:1111")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), testAPIToken) {
		t.Fatal("token must not appear in response body")
	}
}

func TestAPIToken_validSucceeds(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	// Idle session → pause returns 400, but auth must pass first.
	w := postPause(h, "Bearer "+testAPIToken, "10.0.0.3:1111")
	if w.Code == http.StatusUnauthorized {
		t.Fatalf("valid token must not return 401; got %d body=%s", w.Code, w.Body.String())
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 (idle pause), got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), testAPIToken) {
		t.Fatal("token must not appear in response body")
	}
}

func TestAPIToken_unsetPreservesLocalDev(t *testing.T) {
	h := handlerWithEnv(t, nil)
	w := postPause(h, "", "10.0.0.4:1111")
	if w.Code == http.StatusUnauthorized {
		t.Fatal("unset ETH_AMM_SIM_API_TOKEN must not require auth")
	}
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 (idle pause), got %d", w.Code)
	}
}

func TestAPIToken_healthRemainsPublic(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	for _, path := range []string{"/healthz", "/readyz"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.RemoteAddr = "10.0.0.5:1111"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code == http.StatusUnauthorized {
			t.Errorf("%s: health must stay public, got 401", path)
		}
		if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
			t.Errorf("%s: unexpected status %d", path, w.Code)
		}
	}
}

func TestAPIToken_neverLogged(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	_ = postPause(h, "Bearer "+testAPIToken, "10.0.0.6:1111")
	_ = postPause(h, "Bearer wrong", "10.0.0.6:1112")
	_ = postPause(h, "", "10.0.0.6:1113")

	if strings.Contains(buf.String(), testAPIToken) {
		t.Fatalf("token must never appear in logs; got %q", buf.String())
	}
}

func TestAPIToken_startupWarnsWhenUnset(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	_ = handlerWithEnv(t, nil)
	out := buf.String()
	if !strings.Contains(out, "ETH_AMM_SIM_API_TOKEN") || !strings.Contains(strings.ToLower(out), "disabled") {
		t.Fatalf("want startup warning that mutation auth is disabled; got %q", out)
	}
	if strings.Contains(out, testAPIToken) {
		t.Fatal("token must not appear in startup logs")
	}
}

// ── Origin gate ──────────────────────────────────────────────────────────────

func TestOrigin_disallowedRejected(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{
		"ETH_AMM_SIM_ALLOWED_ORIGINS": "https://eth-amm-sim.vercel.app",
	})
	w := getSessionState(h, "https://evil.example", "10.0.1.1:2222")
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Fatal("must not restore wildcard CORS when allowlist is configured")
	}
}

func TestOrigin_allowedSucceeds(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{
		"ETH_AMM_SIM_ALLOWED_ORIGINS": "https://eth-amm-sim.vercel.app",
	})
	w := getSessionState(h, "https://eth-amm-sim.vercel.app", "10.0.1.2:2222")
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://eth-amm-sim.vercel.app" {
		t.Fatalf("Access-Control-Allow-Origin: want allowlisted origin, got %q", got)
	}
}

func TestOrigin_noOriginAllowedWhenAllowlistSet(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{
		"ETH_AMM_SIM_ALLOWED_ORIGINS": "https://eth-amm-sim.vercel.app",
	})
	w := getSessionState(h, "", "10.0.1.3:2222")
	if w.Code != http.StatusOK {
		t.Fatalf("server-to-server (no Origin) must succeed; got %d", w.Code)
	}
}

// ── Rate limit ───────────────────────────────────────────────────────────────

func TestRateLimit_returns429(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{
		"ETH_AMM_SIM_RATE_LIMIT_READ":     "3",
		"ETH_AMM_SIM_RATE_LIMIT_MUTATION": "2",
	})
	addr := "10.0.2.1:3333"
	var last *httptest.ResponseRecorder
	for i := 0; i < 4; i++ {
		last = getSessionState(h, "", addr)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429 after exceeding read limit, got %d", last.Code)
	}
	if ra := last.Header().Get("Retry-After"); ra == "" {
		t.Fatal("429 must include Retry-After")
	}

	mutAddr := "10.0.2.2:3333"
	var mutLast *httptest.ResponseRecorder
	for i := 0; i < 3; i++ {
		mutLast = postPause(h, "", mutAddr)
	}
	if mutLast.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429 after exceeding mutation limit, got %d", mutLast.Code)
	}
	if ra := mutLast.Header().Get("Retry-After"); ra == "" {
		t.Fatal("mutation 429 must include Retry-After")
	}
}

// ── WebSocket auth (subprotocol, not query string) ───────────────────────────

func TestWSAuth_missingToken401(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/stream"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("expected dial failure without token")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		t.Fatalf("want HTTP 401 before upgrade, got %d err=%v", code, err)
	}
}

func TestWSAuth_validSubprotocolSucceeds(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/stream"
	hdr := http.Header{}
	hdr.Set("Sec-WebSocket-Protocol", "eth-amm-sim.bearer."+testAPIToken)
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
	if err != nil {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		t.Fatalf("valid subprotocol must connect; status=%d err=%v", code, err)
	}
	defer conn.Close()
	if conn.Subprotocol() != "eth-amm-sim.bearer."+testAPIToken {
		t.Fatalf("want negotiated subprotocol, got %q", conn.Subprotocol())
	}
}

func TestWSAuth_rejectsQueryToken(t *testing.T) {
	h := handlerWithEnv(t, map[string]string{"ETH_AMM_SIM_API_TOKEN": testAPIToken})
	srv := httptest.NewServer(h)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/stream?token=" + testAPIToken
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("query-string token must not authenticate")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		t.Fatalf("want 401 when only query token present, got %d", code)
	}
}
