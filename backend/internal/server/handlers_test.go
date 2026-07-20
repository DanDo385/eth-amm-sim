package server

import (
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"eth-amm-sim/internal/config"
)

func TestParseResetMode(t *testing.T) {
	tests := []struct {
		name       string
		query      url.Values
		wantMode   string
		wantHard   bool
		wantReseed bool
		wantErr    bool
	}{
		{
			name:       "default soft",
			query:      url.Values{},
			wantMode:   "soft",
			wantHard:   false,
			wantReseed: false,
		},
		{
			name:       "explicit soft",
			query:      url.Values{"mode": []string{"soft"}},
			wantMode:   "soft",
			wantHard:   false,
			wantReseed: false,
		},
		{
			name:       "explicit hard",
			query:      url.Values{"mode": []string{"hard"}},
			wantMode:   "hard",
			wantHard:   true,
			wantReseed: false,
		},
		{
			name:       "explicit reseed",
			query:      url.Values{"mode": []string{"reseed"}},
			wantMode:   "reseed",
			wantHard:   true,
			wantReseed: true,
		},
		{
			name:       "legacy hard flag",
			query:      url.Values{"hard": []string{"true"}},
			wantMode:   "hard",
			wantHard:   true,
			wantReseed: false,
		},
		{
			name:       "legacy reseed flag",
			query:      url.Values{"reseed": []string{"true"}},
			wantMode:   "reseed",
			wantHard:   true,
			wantReseed: true,
		},
		{
			name:       "legacy reseed overrides hard",
			query:      url.Values{"hard": []string{"true"}, "reseed": []string{"true"}},
			wantMode:   "reseed",
			wantHard:   true,
			wantReseed: true,
		},
		{
			name:       "explicit mode ignores legacy flags",
			query:      url.Values{"mode": []string{"hard"}, "reseed": []string{"true"}},
			wantMode:   "hard",
			wantHard:   true,
			wantReseed: false,
		},
		{
			name:    "invalid mode",
			query:   url.Values{"mode": []string{"all-the-things"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, hard, reseed, err := parseResetMode(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (mode=%s hard=%v reseed=%v)", mode, hard, reseed)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mode != tt.wantMode || hard != tt.wantHard || reseed != tt.wantReseed {
				t.Fatalf("unexpected parse result: got (mode=%s hard=%v reseed=%v) want (mode=%s hard=%v reseed=%v)",
					mode, hard, reseed, tt.wantMode, tt.wantHard, tt.wantReseed)
			}
		})
	}
}

func TestEnvironWithFoundryPathPrependsBin(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "forge")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FOUNDRY_BIN", dir)
	t.Setenv("HOME", filepath.Join(dir, "no-home"))

	env := environWithFoundryPath([]string{"FOO=1", "PATH=/usr/bin"})
	var path string
	for _, kv := range env {
		if strings.HasPrefix(kv, "PATH=") {
			path = strings.TrimPrefix(kv, "PATH=")
		}
	}
	if path == "" {
		t.Fatal("expected PATH in env")
	}
	if !strings.HasPrefix(path, dir+":") && path != dir {
		t.Fatalf("expected Foundry bin first on PATH, got %q", path)
	}
	if !strings.Contains(path, "/usr/bin") {
		t.Fatalf("expected original PATH retained, got %q", path)
	}
}

func TestIsAnvilDefaultUserLeak(t *testing.T) {
	toWei := func(n int64) *big.Int {
		return new(big.Int).Mul(big.NewInt(n), big.NewInt(1e18))
	}

	tests := []struct {
		name    string
		eth     *big.Int
		appl    *big.Int
		wantHit bool
	}{
		{name: "exact leak signature", eth: toWei(30000), appl: toWei(config.UserStartingAPPL), wantHit: true},
		{name: "expected normalized balance", eth: toWei(config.UserStartingETH), appl: toWei(config.UserStartingAPPL), wantHit: false},
		{name: "other mismatch", eth: toWei(30000), appl: toWei(900), wantHit: false},
		{name: "nil balances", eth: nil, appl: nil, wantHit: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAnvilDefaultUserLeak(tt.eth, tt.appl)
			if got != tt.wantHit {
				t.Fatalf("unexpected leak detection: got %v want %v", got, tt.wantHit)
			}
		})
	}
}
