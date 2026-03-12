package wf21

import (
	"context"
	"testing"
	"time"
)

func TestSummarySendProbe(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.DryRun {
		t.Fatalf("WF21_DRY_RUN=true; set WF21_DRY_RUN=false for send probe")
	}
	if !cfg.SummarySendEnabled {
		t.Fatalf("WF21_SUMMARY_SEND_ENABLED=false; set true for send probe")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	_, sheetsSvc, err := newGoogleServices(ctx, cfg)
	if err != nil {
		t.Fatalf("init google services: %v", err)
	}

	t.Logf(
		"probe start mode=%s sheet=%s tab=%q range=%s second_enabled=%t second_tab=%q second_ranges=%q extra_enabled=%t extra_count=%d target_group=%s target_groups=%q",
		cfg.SummarySeaTalkMode,
		cfg.SummarySheetID,
		cfg.SummaryTab,
		cfg.SummaryRange,
		cfg.SummarySecondEnabled,
		cfg.SummarySecondTab,
		cfg.SummarySecondRanges,
		cfg.SummaryExtraEnabled,
		len(cfg.SummaryExtraImages),
		cfg.SummarySeaTalkGroupID,
		cfg.SummarySeaTalkGroupIDs,
	)

	result, err := sendSummarySnapshotToSeaTalk(ctx, cfg, sheetsSvc)
	if err != nil {
		t.Fatalf("send summary snapshot: %v", err)
	}
	t.Logf("probe success stable=%t format=%s raw_bytes=%d", result.Stable, result.Format, result.RawBytes)
}
