package metrics

import "testing"

func TestEWMAUpdateAndReset(t *testing.T) {
	e := NewEWMAState(50)
	e.UpdateEWMA(1.0)
	e.UpdateEWMA(2.0)
	if e.GetTradeCount() != 2 {
		t.Fatalf("trade count: got %d", e.GetTradeCount())
	}
	if e.GetSigma() < 0 {
		t.Fatalf("sigma must be non-negative")
	}
	e.ResetEWMA()
	if e.GetTradeCount() != 0 || e.GetMu() != 0 || e.GetSigma() != 0 {
		t.Fatalf("after reset: count=%d mu=%g sigma=%g", e.GetTradeCount(), e.GetMu(), e.GetSigma())
	}
}

func TestEWMAZScoreZeroSigma(t *testing.T) {
	e := NewEWMAState(10)
	if z := e.GetZScore(1.0); z != 0 {
		t.Fatalf("expected 0 z-score before variance, got %v", z)
	}
}
