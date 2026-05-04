package engine

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type testBot struct {
	n       string
	stopped atomic.Bool
}

func (b *testBot) Run(ctx context.Context) {
	<-ctx.Done()
	b.stopped.Store(true)
}

func (b *testBot) Stop() {}

func (b *testBot) Nickname() string { return b.n }

func TestOrchestratorStartStopCancelsBots(t *testing.T) {
	o := NewOrchestrator()
	b := &testBot{n: "test-bot-1"}
	o.AddBot(b)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o.Start(ctx)
	o.Stop()

	time.Sleep(30 * time.Millisecond)
	if !b.stopped.Load() {
		t.Fatal("expected bot Run to observe context cancellation after Stop")
	}
}
