package server

import (
	"testing"
)

func TestCollector_Collect(t *testing.T) {
	c := NewCollector("/tmp")
	s := c.Collect()

	if s.Hostname == "" {
		t.Error("expected non-empty hostname")
	}
	if s.ProcessCount <= 0 {
		t.Errorf("expected positive process count, got %d", s.ProcessCount)
	}
	if s.RecordedAt.IsZero() {
		t.Error("expected non-zero recorded_at")
	}
	if s.Uptime <= 0 {
		t.Error("expected positive uptime")
	}
	if s.RAMTotal == 0 {
		t.Error("expected non-zero RAM total")
	}
	if s.DiskTotal == 0 {
		t.Error("expected non-zero disk total")
	}

	// Second collection should give CPU usage
	s2 := c.Collect()
	t.Logf("CPU usage: %.2f%%, RAM: %.1f%%, Disk: %.1f%%",
		s2.CPUUsage, s2.RAMUsage, s2.DiskUsage)

	if s2.LoadAvg1 == 0 && s2.LoadAvg5 == 0 {
		t.Log("load average is zero — may be expected on lightweight systems")
	}
}
