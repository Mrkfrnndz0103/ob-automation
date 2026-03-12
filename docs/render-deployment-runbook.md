# Render Deployment Runbook (No Blueprint) + Production Env Checklist

## 1. What To Deploy

Deploy this repo as a **Render Web Service** (free tier), connected to GitHub.

Why web service (not cron/serverless):
- This workflow is a long-running poller (`WF21_CONTINUOUS=true`).
- It needs an always-running process with health endpoint (`/healthz`).

## 2. Pre-Deploy Checks (Local)

Run once before pushing:

```bash
go test ./internal/wf21 -run "TestBuildSummarySnapshotTargets|TestBuildSummaryCaption|TestFormatSummarySyncTimestamp|TestSelectPendingZipFiles|TestParseDriveTimestamp"
go build ./...
```

## 3. Create Render Service (Manual, No Blueprint)

1. Push latest code to GitHub.
2. In Render: `New` -> `Web Service` -> connect your GitHub repo.
3. Use these settings:
   - Runtime: `Go`
   - Build Command: `go build -o wf21 ./cmd`
   - Start Command: `./wf21`
   - Auto Deploy: `On` (recommended)
4. Health Check Path: `/healthz`
5. Instance: Free tier

## 4. Production Env Checklist (Important)

Set these in Render environment variables.

### Required credentials

- `WF21_GOOGLE_CREDENTIALS_JSON` (recommended on Render), or file-based creds path if you handle file injection
- `WF21_R2_ACCOUNT_ID`
- `WF21_R2_BUCKET`
- `WF21_R2_ACCESS_KEY_ID`
- `WF21_R2_SECRET_ACCESS_KEY`

### Core workflow

- `WF21_CONTINUOUS=true`
- `WF21_POLL_INTERVAL_SECONDS=1`
- `WF21_DRY_RUN=false`
- `WF21_BOOTSTRAP_PROCESS_EXISTING=true`
- `WF21_ENABLE_HEALTH_SERVER=true`
- Do not hardcode `PORT`; Render injects it automatically (app reads `WF21_HEALTH_PORT` then `PORT`)

### Strongly recommended for stability on Render

Use R2-backed state/lock/status (do **not** rely on local `data/*.json` in production):

- `WF21_STATE_FILE=r2://wf21/state.json`
- `WF21_STATUS_FILE=r2://wf21/status.json`
- `WF21_LOCK_FILE=r2://wf21/lock.json`
- `WF21_LOCK_STALE_AFTER_SECONDS=1200`

Why: Render filesystem is ephemeral; local state loss can cause repeated imports after restart.

### Destination tabs

- `WF21_DESTINATION_TAB_PENDING_RCV=pending_rcv`
- `WF21_DESTINATION_TAB_PACKED_IN_ANOTHER_TO=packed_in_another_to`
- `WF21_DESTINATION_TAB_NO_LHPACKING=no_lhpacking`

### Summary send

- `WF21_SUMMARY_SEND_ENABLED=true`
- `WF21_SUMMARY_SYNC_CELL=config!B1`
- `WF21_SUMMARY_WAIT_SECONDS=8`
- `WF21_SUMMARY_STABILITY_RUNS=3`
- `WF21_SUMMARY_STABILITY_WAIT_SECONDS=1`
- SeaTalk bot mode:
  - `WF21_SUMMARY_SEATALK_MODE=bot`
  - `WF21_SEATALK_APP_ID`
  - `WF21_SEATALK_APP_SECRET`
  - `WF21_SEATALK_GROUP_ID` (or `WF21_SEATALK_GROUP_IDS`)

### PDF renderer choice

If you deploy with native Go runtime on Render free tier, safest is:
- `WF21_SUMMARY_RENDER_MODE=styled`

Reason:
- `pdf_png` mode requires `pdftoppm` or `magick`.
- If those binaries are unavailable, startup/config validation can fail.

If you must use `pdf_png`, use a Docker deployment that explicitly installs Poppler/ImageMagick.

## 5. UptimeRobot Setup

1. Create an HTTP(s) monitor.
2. URL: `https://<your-render-service>.onrender.com/healthz`
3. Interval: `5 minutes` (free plan).
4. Add alert contacts (email/Telegram/Slack as needed).

Optional second monitor:
- URL: `/status`
- Use only if `WF21_STATUS_FILE` is configured.

## 6. Post-Deploy Verification

Check Render logs for:
- `watch mode enabled poll_interval=1s`
- `health server listening on ...`
- no repeating `cycle error=...`

Validate endpoints:

```bash
curl https://<service>.onrender.com/healthz
curl https://<service>.onrender.com/status
```

Expected:
- `/healthz` returns `ok`
- `/status` returns JSON with `changed=false` when no new ZIP exists

## 6.1 Verify `pdftoppm` / `magick` On Render Host

You cannot inspect Render host binaries locally; verify from Render deploy/runtime behavior:

1. Set:
   - `WF21_SUMMARY_RENDER_MODE=pdf_png`
   - `WF21_SUMMARY_PDF_CONVERTER=auto` (or `pdftoppm` / `magick`)
2. Deploy and check Render logs.
3. If converter is missing, startup fails with config error similar to:
   - `WF21_SUMMARY_RENDER_MODE=pdf_png requires converter availability: ...`

Optional temporary Start Command probe (for one deploy):

```bash
sh -lc 'which pdftoppm || true; which magick || true; ./wf21'
```

Then revert Start Command back to:

```bash
./wf21
```

## 7. Troubleshooting Quick Notes

- Re-import happens twice:
  - Ensure only one running instance.
  - Ensure lock is configured (`WF21_LOCK_FILE`).
  - Ensure state is persisted on R2 (`WF21_STATE_FILE=r2://...`), not local disk.
- Summary image not sent:
  - Check summary stability settings.
  - Check SeaTalk credentials/group IDs.
  - If using `pdf_png`, verify converter binaries exist.

## Sources

- Render Go deploy docs: https://render.com/docs/deploy-go
- Render health checks: https://render.com/docs/health-checks
- Render free tier behavior: https://render.com/free
- Render persistent disks: https://render.com/docs/disks
- UptimeRobot interval docs: https://help.uptimerobot.com/en/articles/11360876-what-is-a-monitoring-interval-in-uptimerobot
