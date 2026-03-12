# Render Deployment Runbook (Docker Required)

## 1. Decision

For this project, deployment **must use Docker**.

Reason:
- You require `WF21_SUMMARY_RENDER_MODE=pdf_png`.
- `pdf_png` requires runtime binaries: `pdftoppm` or `magick`.
- Docker guarantees these tools are always present in every deploy.

Implemented in repo:
- `Dockerfile` installs `poppler-utils` (`pdftoppm`) and `imagemagick` (`magick`)
- `.dockerignore` reduces build context and excludes secrets/logs

## 2. Pre-Deploy (Local)

Run before pushing:

```bash
go test ./internal/wf21 -run "TestBuildSummarySnapshotTargets|TestBuildSummaryCaption|TestFormatSummarySyncTimestamp|TestSelectPendingZipFiles|TestParseDriveTimestamp"
go build ./...
```

Optional local container check:

```bash
docker build -t wf21:local .
docker run --rm wf21:local sh -lc "which pdftoppm && which magick"
```

## 3. Create Render Service (Docker, No Blueprint)

1. Push code to GitHub (including `Dockerfile`).
2. In Render: `New` -> `Web Service` -> connect repo.
3. Render will detect Docker automatically.
4. Use:
   - Environment: Docker
   - Dockerfile Path: `./Dockerfile`
   - Auto Deploy: `On`
5. Health Check Path: `/healthz`
6. Instance: Free tier

Do not set Go build/start commands in Render for this setup.

## 4. Required Environment Variables (Render)

### Credentials

- `WF21_GOOGLE_CREDENTIALS_JSON`
- `WF21_R2_ACCOUNT_ID`
- `WF21_R2_BUCKET`
- `WF21_R2_ACCESS_KEY_ID`
- `WF21_R2_SECRET_ACCESS_KEY`

### Workflow core

- `WF21_CONTINUOUS=true`
- `WF21_POLL_INTERVAL_SECONDS=1`
- `WF21_DRY_RUN=false`
- `WF21_BOOTSTRAP_PROCESS_EXISTING=true`
- `WF21_ENABLE_HEALTH_SERVER=true`

Do not hardcode `PORT`; Render injects it.

### State, status, lock (strongly recommended)

Use R2-backed paths to survive restarts/redeploys:

- `WF21_STATE_FILE=r2://wf21/state.json`
- `WF21_STATUS_FILE=r2://wf21/status.json`
- `WF21_LOCK_FILE=r2://wf21/lock.json`
- `WF21_LOCK_STALE_AFTER_SECONDS=1200`

### Destination tabs

- `WF21_DESTINATION_TAB_PENDING_RCV=pending_rcv`
- `WF21_DESTINATION_TAB_PACKED_IN_ANOTHER_TO=packed_in_another_to`
- `WF21_DESTINATION_TAB_NO_LHPACKING=no_lhpacking`

### Summary + SeaTalk

- `WF21_SUMMARY_SEND_ENABLED=true`
- `WF21_SUMMARY_SYNC_CELL=config!B1`
- `WF21_SUMMARY_WAIT_SECONDS=8`
- `WF21_SUMMARY_STABILITY_RUNS=3`
- `WF21_SUMMARY_STABILITY_WAIT_SECONDS=1`
- `WF21_SUMMARY_SEATALK_MODE=bot`
- `WF21_SEATALK_APP_ID`
- `WF21_SEATALK_APP_SECRET`
- `WF21_SEATALK_GROUP_ID` (or `WF21_SEATALK_GROUP_IDS`)

### Renderer (required by your decision)

- `WF21_SUMMARY_RENDER_MODE=pdf_png`
- `WF21_SUMMARY_PDF_CONVERTER=auto` (or `pdftoppm`)

## 5. Verify Converter Availability On Render

Check Render logs after deploy. You should not see:

- `WF21_SUMMARY_RENDER_MODE=pdf_png requires converter availability: ...`

Optional one-time start probe:

```bash
sh -lc 'which pdftoppm; which magick; ./wf21'
```

Then revert to default Docker CMD.

## 6. UptimeRobot Setup

1. Create HTTP(s) monitor.
2. URL: `https://<your-service>.onrender.com/healthz`
3. Interval: `5 minutes` (free plan).

Optional second monitor:
- `https://<your-service>.onrender.com/status`

## 7. Post-Deploy Verification

Expected logs:
- `watch mode enabled poll_interval=1s`
- `health server listening on ...`

Expected endpoints:
- `/healthz` -> `ok`
- `/status` -> JSON; when no new zip, `changed=false`

## 8. Troubleshooting

- Duplicate import:
  - Ensure only one service instance.
  - Ensure `WF21_LOCK_FILE` is set.
  - Ensure state file uses `r2://...`, not local disk.
- Summary send missing:
  - Check SeaTalk credentials/group IDs.
  - Check summary stability settings.
  - Check logs for `post-import warning`.

## Sources

- Render Docker deploy docs: https://render.com/docs/docker
- Render health checks: https://render.com/docs/health-checks
- Render free tier behavior: https://render.com/free
- Render persistent disks: https://render.com/docs/disks
- UptimeRobot interval docs: https://help.uptimerobot.com/en/articles/11360876-what-is-a-monitoring-interval-in-uptimerobot
