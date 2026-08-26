// config_test.go - Broadcast JSON address loading (run-latest preference).
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAddressesFromBroadcast_UsesRunLatest(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "backend"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts", "broadcast", "Deploy.s.sol", "31337"), 0o755); err != nil {
		t.Fatal(err)
	}

	json := `{"transactions":[
		{"contractName":"AppleToken","contractAddress":"0x0000000000000000000000000000000000000001"},
		{"contractName":"AppleAMM","contractAddress":"0x0000000000000000000000000000000000000002"}
	]}`
	p := filepath.Join(root, "contracts", "broadcast", "Deploy.s.sol", "31337", "run-latest.json")
	if err := os.WriteFile(p, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(filepath.Join(root, "backend"))

	token, amm, err := LoadAddressesFromBroadcast()
	if err != nil {
		t.Fatalf("LoadAddressesFromBroadcast: %v", err)
	}
	if token != "0x0000000000000000000000000000000000000001" {
		t.Fatalf("token: got %q", token)
	}
	if amm != "0x0000000000000000000000000000000000000002" {
		t.Fatalf("amm: got %q", amm)
	}
}
