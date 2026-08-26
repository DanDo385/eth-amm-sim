// client_test.go - Anvil JSON-RPC helpers (setBalance payload shape).
package chain

import (
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSetBalance_SendsAnvilRPCPayload(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":true}`))
	}))
	t.Cleanup(srv.Close)

	c := &Client{rpcURL: srv.URL}
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	if err := c.SetBalance(context.Background(), addr, big.NewInt(123)); err != nil {
		t.Fatalf("SetBalance: %v", err)
	}
	if got["method"] != "anvil_setBalance" {
		t.Fatalf("method: got %v", got["method"])
	}
}
