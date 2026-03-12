## wf21-drive-csv-consolidation (standalone)

This repository contains only **WF2.1 Drive CSV Consolidation**.

## Prerequisites

- Go 1.22+

## Run

1. Copy `.env.example` to `.env` and fill required values.
2. Install dependencies:

```bash
go mod tidy
```

3. Start:

```bash
go run ./cmd
```

## Workflow

`workflow_2_1_drive_csv_consolidation` does:

1. Polls a Google Drive parent folder for new `.zip` files and processes pending uploads oldest to newest.
2. Reads all `.csv` files from each zip and consolidates into one CSV.
3. Keeps the first CSV header as canonical and aligns subsequent rows by header name.
4. Detects and drops hidden leading unnamed column (enabled by default).
5. Uploads consolidated CSV to Cloudflare R2.
6. Overwrites destination columns `A:K` (keeps `L+` formulas) and imports rows into three tabs in batches.
7. Updates summary sync cell, renders summary image(s), then sends caption + images to SeaTalk.

Default targets:

- Drive parent folder: `1oU9kj5VIJIoNrR388wYCHSdtHGanRrgZ`
- Destination sheet: `1mdi-8ACluDHGZ7yAyNLwXLwpmQ4f6VAx3kpbaJORViA`
- Destination tabs:
  - `pending_rcv`
  - `packed_in_another_to`
  - `no_lhpacking`

## Required Environment Variables

- Google credentials (one of):
  - `WF21_GOOGLE_CREDENTIALS_FILE` (or `GOOGLE_APPLICATION_CREDENTIALS`)
  - `WF21_GOOGLE_CREDENTIALS_JSON`
- `WF21_R2_ACCOUNT_ID`
- `WF21_R2_BUCKET`
- `WF21_R2_ACCESS_KEY_ID`
- `WF21_R2_SECRET_ACCESS_KEY`
- SeaTalk summary send (when `WF21_SUMMARY_SEND_ENABLED=true`):
  - `WF21_SUMMARY_SEATALK_MODE=bot` with:
    - `WF21_SEATALK_GROUP_ID` (or `WF21_SEATALK_GROUP_IDS`)
    - `WF21_SEATALK_APP_ID`
    - `WF21_SEATALK_APP_SECRET`
  - or `WF21_SUMMARY_SEATALK_MODE=webhook` with:
    - `WF21_SEATALK_WEBHOOK_URL` (or `SEATALK_SYSTEM_WEBHOOK_URL`)

## WF21 Optional Environment Variables

- Core flow:
  - `WF21_DRIVE_PARENT_FOLDER_ID`
  - `WF21_DESTINATION_SHEET_ID`
  - `WF21_DESTINATION_TAB_PENDING_RCV`
  - `WF21_DESTINATION_TAB_PACKED_IN_ANOTHER_TO`
  - `WF21_DESTINATION_TAB_NO_LHPACKING`
  - `WF21_R2_OBJECT_PREFIX`
  - `WF21_STATE_FILE`
  - `WF21_STATUS_FILE`
  - `WF21_LOCK_FILE`
  - `WF21_LOCK_STALE_AFTER_SECONDS`
  - `WF21_BOOTSTRAP_PROCESS_EXISTING`
  - `WF21_DROP_LEADING_UNNAMED_COLUMN`
  - `WF21_DRY_RUN`
  - `WF21_CONTINUOUS`
  - `WF21_POLL_INTERVAL_SECONDS`
  - `WF21_SHEETS_BATCH_SIZE`
  - `WF21_SHEETS_WRITE_RETRY_MAX_ATTEMPTS`
  - `WF21_SHEETS_WRITE_RETRY_BASE_MS`
  - `WF21_SHEETS_WRITE_RETRY_MAX_MS`
  - `WF21_TEMP_DIR`
- Summary send/render:
  - `WF21_SUMMARY_SEND_ENABLED`
  - `WF21_SUMMARY_SHEET_ID`
  - `WF21_SUMMARY_TAB`
  - `WF21_SUMMARY_RANGE`
  - `WF21_SUMMARY_SECOND_IMAGE_ENABLED`
  - `WF21_SUMMARY_SECOND_TAB`
  - `WF21_SUMMARY_SECOND_RANGES`
  - `WF21_SUMMARY_EXTRA_IMAGES_ENABLED`
  - `WF21_SUMMARY_EXTRA_IMAGES`
  - `WF21_SUMMARY_SYNC_CELL`
  - `WF21_SUMMARY_WAIT_SECONDS`
  - `WF21_SUMMARY_STABILITY_RUNS`
  - `WF21_SUMMARY_STABILITY_WAIT_SECONDS`
  - `WF21_SUMMARY_RENDER_MODE`
  - `WF21_SUMMARY_RENDER_SCALE`
  - `WF21_SUMMARY_AUTO_FIT_COLUMNS`
  - `WF21_SUMMARY_PDF_DPI`
  - `WF21_SUMMARY_PDF_CONVERTER`
  - `WF21_SUMMARY_IMAGE_MAX_WIDTH_PX`
  - `WF21_SUMMARY_IMAGE_MAX_BASE64_BYTES`
  - `WF21_SUMMARY_HTTP_TIMEOUT_SECONDS`
  - `WF21_TIMEZONE`
- SeaTalk bot/webhook source:
  - `WF21_SEATALK_BASE_URL`
  - `WF21_SEATALK_GROUP_IDS`

## PDF Render Dependency

When `WF21_SUMMARY_RENDER_MODE=pdf_png`, install one of:

- Poppler (`pdftoppm` in PATH), or
- ImageMagick (`magick` in PATH)

## Health Endpoints

When `WF21_ENABLE_HEALTH_SERVER=true`:

- `GET /healthz`
- `GET /status` (only if `WF21_STATUS_FILE` is enabled)

## Security Notes

- Keep all secrets and webhook URLs private.
- Do not share `.env` values, access keys, or credential JSON.
