package wf21

import (
	"strings"
	"testing"
	"time"
)

func TestBuildSummaryCaption_Format(t *testing.T) {
	ts := time.Date(2026, 3, 12, 15, 38, 0, 0, time.UTC)
	got := buildSummaryCaption(ts)
	want := "@All\nOutbound Pending for Dispatch as of 3:38 PM Mar-12. Thanks!"
	if got != want {
		t.Fatalf("unexpected caption: got=%q want=%q", got, want)
	}
}

func TestBuildSummaryCaptionForBot_IncludesMentionAndAtAll(t *testing.T) {
	ts := time.Date(2026, 3, 12, 15, 38, 0, 0, time.UTC)
	got := buildSummaryCaptionForBot(ts)
	if !strings.Contains(got, "<mention-tag target=\"seatalk://user?id=0\"/>") {
		t.Fatalf("expected mention-tag in bot caption, got=%q", got)
	}
	if !strings.Contains(got, "@All\nOutbound Pending for Dispatch as of 3:38 PM Mar-12. Thanks!") {
		t.Fatalf("expected @All summary body in bot caption, got=%q", got)
	}
}

