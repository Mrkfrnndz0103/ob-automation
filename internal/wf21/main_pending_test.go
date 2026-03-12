package wf21

import (
	"testing"
	"time"
)

func TestSelectPendingZipFiles_UsesListOrderAfterLastProcessed(t *testing.T) {
	base := time.Date(2026, 3, 12, 7, 13, 41, 0, time.UTC)
	files := []driveZipFile{
		{ID: "z-last-in-lex", Name: "a.zip", ModifiedTime: base, CreatedTime: base.Add(1 * time.Second)},
		{ID: "m-processed", Name: "b.zip", ModifiedTime: base, CreatedTime: base.Add(2 * time.Second)},
		{ID: "a-new-in-lex", Name: "c.zip", ModifiedTime: base, CreatedTime: base.Add(3 * time.Second)},
	}
	state := workflowState{
		LastProcessedFileID:       "m-processed",
		LastProcessedFileMD5:      "same",
		LastProcessedModifiedTime: base.Format(time.RFC3339Nano),
	}
	files[1].MD5Checksum = "same"

	pending := selectPendingZipFiles(files, state)
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending file, got %d", len(pending))
	}
	if pending[0].ID != "a-new-in-lex" {
		t.Fatalf("expected pending id a-new-in-lex, got %s", pending[0].ID)
	}
}

func TestSelectPendingZipFiles_SkipsSameIDEvenWhenMD5Changed(t *testing.T) {
	base := time.Date(2026, 3, 12, 7, 13, 41, 0, time.UTC)
	files := []driveZipFile{
		{ID: "same-id", Name: "same.zip", MD5Checksum: "new-md5", ModifiedTime: base},
	}
	state := workflowState{
		LastProcessedFileID:       "same-id",
		LastProcessedFileMD5:      "old-md5",
		LastProcessedModifiedTime: base.Format(time.RFC3339Nano),
	}

	pending := selectPendingZipFiles(files, state)
	if len(pending) != 0 {
		t.Fatalf("expected no pending files for same id, got %d", len(pending))
	}
}

func TestSelectPendingZipFiles_LastIDMissingFallsBackToNewerModified(t *testing.T) {
	base := time.Date(2026, 3, 12, 7, 13, 41, 0, time.UTC)
	files := []driveZipFile{
		{ID: "old-a", Name: "old-a.zip", ModifiedTime: base},
		{ID: "old-b", Name: "old-b.zip", ModifiedTime: base},
		{ID: "new-c", Name: "new-c.zip", ModifiedTime: base.Add(1 * time.Minute)},
	}
	state := workflowState{
		LastProcessedFileID:       "deleted-id",
		LastProcessedModifiedTime: base.Format(time.RFC3339Nano),
	}

	pending := selectPendingZipFiles(files, state)
	if len(pending) != 1 {
		t.Fatalf("expected 1 newer pending file, got %d", len(pending))
	}
	if pending[0].ID != "new-c" {
		t.Fatalf("expected pending id new-c, got %s", pending[0].ID)
	}
}

func TestParseDriveTimestamp_AcceptsFractionalRFC3339(t *testing.T) {
	ts := parseDriveTimestamp("2026-03-12T07:13:41.123456Z")
	if ts.IsZero() {
		t.Fatal("expected non-zero timestamp for RFC3339 with fractional seconds")
	}
}
