package wf21

import (
	"testing"
	"time"
)

func TestFormatSummarySyncTimestamp_MMDDHHMM(t *testing.T) {
	ts := time.Date(2026, 3, 12, 15, 7, 59, 0, time.UTC)
	got := formatSummarySyncTimestamp(ts)
	if got != "03-12 15:07" {
		t.Fatalf("unexpected format: got=%q want=%q", got, "03-12 15:07")
	}
}

