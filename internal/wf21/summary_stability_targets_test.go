package wf21

import "testing"

func TestBuildSummarySnapshotTargets_PrimaryAndSecondary(t *testing.T) {
	cfg := workflowConfig{
		SummaryTab:           "[SOC] Backlogs Summary",
		SummaryRange:         "B2:Q59",
		SummarySecondEnabled: true,
		SummarySecondTab:     "config",
		SummarySecondRanges:  []string{"E157:Y195", "B202:T228"},
	}

	targets := buildSummarySnapshotTargets(cfg)
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[0].Tab != "[SOC] Backlogs Summary" || len(targets[0].Ranges) != 1 || targets[0].Ranges[0] != "B2:Q59" {
		t.Fatalf("unexpected primary target: %+v", targets[0])
	}
	if targets[1].Tab != "config" || len(targets[1].Ranges) != 2 {
		t.Fatalf("unexpected secondary target: %+v", targets[1])
	}
	if targets[1].Ranges[0] != "E157:Y195" || targets[1].Ranges[1] != "B202:T228" {
		t.Fatalf("unexpected secondary ranges: %+v", targets[1].Ranges)
	}
}

func TestBuildSummarySnapshotTargets_SecondaryDisabled(t *testing.T) {
	cfg := workflowConfig{
		SummaryTab:           "[SOC] Backlogs Summary",
		SummaryRange:         "B2:Q59",
		SummarySecondEnabled: false,
		SummarySecondTab:     "config",
		SummarySecondRanges:  []string{"E157:Y195", "B202:T228"},
	}

	targets := buildSummarySnapshotTargets(cfg)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
}

