# Cloud Run Jobs Deployment Runbook (Docker Required)

## 1. Decision

For this project, use **Cloud Run Jobs** (not Cloud Run services).

Why this is the right fit for this repo:
- The app already supports one-shot execution via `WF21_CONTINUOUS=false`.
- Your Docker image already includes required PDF converters (`pdftoppm`, `magick`).
- A scheduled job avoids always-on free-tier hour limits from web-service platforms.

This runbook is written for **PowerShell on Windows** and this repository layout.

## 2. Prerequisites

1. Google Cloud project with an **active billing account**.
2. Installed and authenticated CLI:
   - `gcloud` (latest stable)
   - Docker (optional if using Cloud Build only)
3. Access to current workflow secrets (Google credentials JSON, R2 keys, SeaTalk credentials).
4. Existing `WF21_STATE_FILE`/`WF21_LOCK_FILE` in R2 must stay consistent to avoid reprocessing.

## 3. Files In This Repo

- Env template for Cloud Run Jobs: `docs/cloud-run-job.env.yaml.example`
- This runbook: `docs/cloud-run-jobs-deployment-runbook.md`

Create your actual env file:

```powershell
Copy-Item docs/cloud-run-job.env.yaml.example data/cloud-run-job.env.yaml
```

Then edit `data/cloud-run-job.env.yaml` and replace placeholders.

Important:
- Keep secrets out of `data/cloud-run-job.env.yaml`.
- Use Secret Manager + `--set-secrets` for sensitive values.

## 4. Migration Cutover Plan (Avoid Duplicate Imports)

1. Keep current Render service running while you prepare Cloud Run.
2. Deploy Cloud Run Job and run a manual test execution.
3. Pause/disable Render service.
4. Enable Cloud Scheduler trigger.

This prevents two runtimes processing the same Drive events at the same time.

## 5. Set Variables (PowerShell)

Run these commands in repo root:

```powershell
$PROJECT_ID = "REPLACE_WITH_PROJECT_ID"
$REGION = "asia-southeast1"
$REPO = "wf21"
$IMAGE_NAME = "wf21-drive-csv-consolidation"
$JOB_NAME = "wf21-drive-csv-consolidation"
$RUNTIME_SA_NAME = "wf21-job-runtime"
$SCHEDULER_SA_NAME = "wf21-scheduler-invoker"
$SCHEDULER_JOB_NAME = "wf21-drive-csv-every-5m"
$SCHEDULER_REGION = "asia-southeast1"
$CRON = "*/5 * * * *"
$TIME_ZONE = "Asia/Singapore"

gcloud config set project $PROJECT_ID
$PROJECT_NUMBER = (gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
$RUNTIME_SA = "$RUNTIME_SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"
$SCHEDULER_SA = "$SCHEDULER_SA_NAME@$PROJECT_ID.iam.gserviceaccount.com"
```

## 6. Enable Required APIs

```powershell
gcloud services enable `
  run.googleapis.com `
  cloudbuild.googleapis.com `
  artifactregistry.googleapis.com `
  cloudscheduler.googleapis.com `
  secretmanager.googleapis.com `
  iam.googleapis.com `
  logging.googleapis.com
```

## 7. Build And Push Docker Image

Create Artifact Registry repo (one-time):

```powershell
gcloud artifacts repositories create $REPO `
  --repository-format=docker `
  --location=$REGION `
  --description="WF21 Docker images"
```

If it already exists, continue.

Build image with Cloud Build:

```powershell
$TAG = Get-Date -Format "yyyyMMdd-HHmmss"
$IMAGE_URI = "$REGION-docker.pkg.dev/$PROJECT_ID/$REPO/$IMAGE_NAME:$TAG"

gcloud builds submit --tag $IMAGE_URI .
```

## 8. Create Service Accounts

Create runtime service account:

```powershell
gcloud iam service-accounts create $RUNTIME_SA_NAME `
  --display-name="WF21 Cloud Run Job Runtime"
```

Create scheduler invoker service account:

```powershell
gcloud iam service-accounts create $SCHEDULER_SA_NAME `
  --display-name="WF21 Scheduler Invoker"
```

If either already exists, continue.

## 9. Create Secrets (One-Time) And Grant Access

Recommended secret names and env mapping:

- `wf21-google-credentials-json` -> `WF21_GOOGLE_CREDENTIALS_JSON`
- `wf21-r2-access-key-id` -> `WF21_R2_ACCESS_KEY_ID`
- `wf21-r2-secret-access-key` -> `WF21_R2_SECRET_ACCESS_KEY`
- `wf21-seatalk-app-id` -> `WF21_SEATALK_APP_ID`
- `wf21-seatalk-app-secret` -> `WF21_SEATALK_APP_SECRET`
- Optional: `wf21-seatalk-webhook-url` -> `WF21_SEATALK_WEBHOOK_URL`
- Optional: `wf21-newrelic-license-key` -> `WF21_NEWRELIC_LICENSE_KEY`

Create secret containers (one-time):

```powershell
gcloud secrets create wf21-google-credentials-json --replication-policy=automatic
gcloud secrets create wf21-r2-access-key-id --replication-policy=automatic
gcloud secrets create wf21-r2-secret-access-key --replication-policy=automatic
gcloud secrets create wf21-seatalk-app-id --replication-policy=automatic
gcloud secrets create wf21-seatalk-app-secret --replication-policy=automatic
```

Add secret versions (repeat any time you rotate values):

```powershell
# Example using local files that are NOT committed.
gcloud secrets versions add wf21-google-credentials-json --data-file=".deploy/secrets/wf21-google-credentials.json"
gcloud secrets versions add wf21-r2-access-key-id --data-file=".deploy/secrets/wf21-r2-access-key-id.txt"
gcloud secrets versions add wf21-r2-secret-access-key --data-file=".deploy/secrets/wf21-r2-secret-access-key.txt"
gcloud secrets versions add wf21-seatalk-app-id --data-file=".deploy/secrets/wf21-seatalk-app-id.txt"
gcloud secrets versions add wf21-seatalk-app-secret --data-file=".deploy/secrets/wf21-seatalk-app-secret.txt"
```

Grant runtime service account secret access:

```powershell
$secrets = @(
  "wf21-google-credentials-json",
  "wf21-r2-access-key-id",
  "wf21-r2-secret-access-key",
  "wf21-seatalk-app-id",
  "wf21-seatalk-app-secret"
)

foreach ($s in $secrets) {
  gcloud secrets add-iam-policy-binding $s `
    --member="serviceAccount:$RUNTIME_SA" `
    --role="roles/secretmanager.secretAccessor"
}
```

## 10. Create Cloud Run Job

Make sure `data/cloud-run-job.env.yaml` is fully updated first.

```powershell
gcloud run jobs create $JOB_NAME `
  --image $IMAGE_URI `
  --region $REGION `
  --service-account $RUNTIME_SA `
  --tasks 1 `
  --parallelism 1 `
  --max-retries 0 `
  --task-timeout 3600s `
  --cpu 1 `
  --memory 1Gi `
  --env-vars-file data/cloud-run-job.env.yaml `
  --set-secrets "WF21_GOOGLE_CREDENTIALS_JSON=wf21-google-credentials-json:latest,WF21_R2_ACCESS_KEY_ID=wf21-r2-access-key-id:latest,WF21_R2_SECRET_ACCESS_KEY=wf21-r2-secret-access-key:latest,WF21_SEATALK_APP_ID=wf21-seatalk-app-id:latest,WF21_SEATALK_APP_SECRET=wf21-seatalk-app-secret:latest"
```

Notes:
- `--max-retries 0` is deliberate to reduce duplicate side effects on partial failures.
- `WF21_ENABLE_HEALTH_SERVER=false` is expected for jobs.
- `WF21_CONTINUOUS=false` is required.

## 11. Manual Test Execution

Run once and wait for completion:

```powershell
gcloud run jobs execute $JOB_NAME --region $REGION --wait
```

Check recent executions:

```powershell
gcloud run jobs executions list --job $JOB_NAME --region $REGION
```

Read logs:

```powershell
gcloud logging read "resource.type=cloud_run_job AND resource.labels.job_name=$JOB_NAME" `
  --limit=100 `
  --format="value(timestamp,textPayload)"
```

Expected behavior:
- If new zip exists: processing/import logs appear.
- If no new zip exists: logs indicate no changes and exit cleanly.

## 12. Authorize Scheduler To Run The Job

Grant Scheduler invoker role on this job:

```powershell
gcloud run jobs add-iam-policy-binding $JOB_NAME `
  --region $REGION `
  --member "serviceAccount:$SCHEDULER_SA" `
  --role "roles/run.invoker"
```

## 13. Create Cloud Scheduler Trigger

Create schedule (every 5 minutes example):

```powershell
$RUN_URI = "https://run.googleapis.com/v2/projects/$PROJECT_ID/locations/$REGION/jobs/$JOB_NAME:run"

gcloud scheduler jobs create http $SCHEDULER_JOB_NAME `
  --location=$SCHEDULER_REGION `
  --schedule="$CRON" `
  --time-zone="$TIME_ZONE" `
  --http-method=POST `
  --uri="$RUN_URI" `
  --oauth-service-account-email="$SCHEDULER_SA" `
  --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform"
```

If the scheduler job already exists, use update:

```powershell
gcloud scheduler jobs update http $SCHEDULER_JOB_NAME `
  --location=$SCHEDULER_REGION `
  --schedule="$CRON" `
  --time-zone="$TIME_ZONE" `
  --http-method=POST `
  --uri="$RUN_URI" `
  --oauth-service-account-email="$SCHEDULER_SA" `
  --oauth-token-scope="https://www.googleapis.com/auth/cloud-platform"
```

## 14. Production Cutover Checklist

1. Confirm one manual execution succeeds.
2. Pause/disable Render service.
3. Enable Cloud Scheduler job.
4. Watch first 2-3 scheduled executions.
5. Confirm state file updates in R2 and no duplicate imports.

## 15. Ongoing Operations

Deploy a new version:

```powershell
$TAG = Get-Date -Format "yyyyMMdd-HHmmss"
$IMAGE_URI = "$REGION-docker.pkg.dev/$PROJECT_ID/$REPO/$IMAGE_NAME:$TAG"

gcloud builds submit --tag $IMAGE_URI .

gcloud run jobs update $JOB_NAME `
  --region $REGION `
  --image $IMAGE_URI
```

Run immediately after deploy:

```powershell
gcloud run jobs execute $JOB_NAME --region $REGION --wait
```

Pause scheduler quickly:

```powershell
gcloud scheduler jobs pause $SCHEDULER_JOB_NAME --location $SCHEDULER_REGION
```

Resume scheduler:

```powershell
gcloud scheduler jobs resume $SCHEDULER_JOB_NAME --location $SCHEDULER_REGION
```

## 16. Cost Guardrails (Recommended)

1. Create a Cloud Billing budget and email alerts.
2. Start with conservative schedule (`*/5 * * * *`), then tighten only if needed.
3. Keep `tasks=1`, `parallelism=1` unless you intentionally redesign for parallel imports.
4. Keep `WF21_STATE_FILE` and `WF21_LOCK_FILE` in R2.

## 17. Troubleshooting

- Error: `requires converter availability`:
  - Ensure deploy uses this repo Dockerfile (contains `poppler-utils` and `imagemagick`).
- Error: Secret access denied:
  - Ensure runtime SA has `roles/secretmanager.secretAccessor` on each secret.
- Error: Scheduler 403 calling Run API:
  - Ensure scheduler SA has `roles/run.invoker` on the Cloud Run Job.
- Duplicate imports:
  - Ensure Render is fully stopped.
  - Ensure one scheduler job only.
  - Ensure lock/state use shared R2 paths.
- No files processed on first run:
  - Check `WF21_BOOTSTRAP_PROCESS_EXISTING` behavior. `false` sets baseline when no state exists.

## 18. Google Cloud Console UI Setup (No CLI)

Use this section if you prefer configuring in the Google Cloud web console.

### 18.1 Select Project And Enable APIs

1. Open Google Cloud Console and select your target project.
2. Go to `APIs & Services` -> `Enabled APIs & services` -> `+ ENABLE APIS AND SERVICES`.
3. Enable:
   - Cloud Run API
   - Cloud Build API
   - Artifact Registry API
   - Cloud Scheduler API
   - Secret Manager API
   - IAM API
   - Cloud Logging API

### 18.2 Create Artifact Registry Docker Repository

1. Go to `Artifact Registry` -> `Repositories` -> `Create Repository`.
2. Set:
   - Name: `wf21`
   - Format: `Docker`
   - Region: `asia-southeast1` (or your chosen region)
3. Click `Create`.

### 18.3 Build Image Using Cloud Build Trigger (UI)

1. Go to `Cloud Build` -> `Triggers` -> `Create Trigger`.
2. Connect your GitHub repo if not connected yet.
3. Configure trigger:
   - Event: manual trigger (or push trigger if you want auto-builds)
   - Configuration: `Dockerfile`
   - Dockerfile location: `Dockerfile`
   - Build context: repository root
   - Image destination: `asia-southeast1-docker.pkg.dev/<PROJECT_ID>/wf21/wf21-drive-csv-consolidation`
4. Save trigger.
5. Click `Run` on the trigger to build and push an image.
6. Copy the resulting image URI from build results.

### 18.4 Create Service Accounts In UI

1. Go to `IAM & Admin` -> `Service Accounts` -> `Create Service Account`.
2. Create runtime account:
   - Name: `wf21-job-runtime`
3. Create scheduler invoker account:
   - Name: `wf21-scheduler-invoker`
4. Keep both emails for later steps.

### 18.5 Create Secrets In Secret Manager (UI)

1. Go to `Security` -> `Secret Manager` -> `Create Secret`.
2. Create these secrets (names should match this runbook):
   - `wf21-google-credentials-json`
   - `wf21-r2-access-key-id`
   - `wf21-r2-secret-access-key`
   - `wf21-seatalk-app-id`
   - `wf21-seatalk-app-secret`
3. For each secret, add its value as secret data (paste or upload file).
4. Open each secret -> `Permissions` -> `Grant Access`:
   - Principal: `wf21-job-runtime@<PROJECT_ID>.iam.gserviceaccount.com`
   - Role: `Secret Manager Secret Accessor`

### 18.6 Create Cloud Run Job In UI

1. Go to `Cloud Run` -> `Jobs` -> `Create Job`.
2. Set:
   - Job name: `wf21-drive-csv-consolidation`
   - Region: `asia-southeast1`
   - Container image URL: image URI built in step 18.3
3. In execution settings:
   - Tasks: `1`
   - Parallelism: `1`
   - Max retries: `0`
   - Timeout: `3600` seconds
   - CPU: `1`
   - Memory: `1 GiB`
4. Set service account to `wf21-job-runtime@<PROJECT_ID>.iam.gserviceaccount.com`.
5. In variables and secrets:
   - Add non-secret env vars from `docs/cloud-run-job.env.yaml.example` (use your values).
   - Add secret-backed env vars:
     - `WF21_GOOGLE_CREDENTIALS_JSON` from `wf21-google-credentials-json`
     - `WF21_R2_ACCESS_KEY_ID` from `wf21-r2-access-key-id`
     - `WF21_R2_SECRET_ACCESS_KEY` from `wf21-r2-secret-access-key`
     - `WF21_SEATALK_APP_ID` from `wf21-seatalk-app-id`
     - `WF21_SEATALK_APP_SECRET` from `wf21-seatalk-app-secret`
6. Confirm these values before deploy:
   - `WF21_CONTINUOUS=false`
   - `WF21_ENABLE_HEALTH_SERVER=false`
7. Click `Create`.

### 18.7 Execute And Verify In UI

1. Open the job page in `Cloud Run` -> `Jobs`.
2. Click `Execute`.
3. Open the execution details and logs.
4. Confirm either:
   - New file processed and imported, or
   - `no new zip files to process` and clean exit.

### 18.8 Grant Scheduler Invoker Permission In UI

1. Open `Cloud Run` -> `Jobs` -> `wf21-drive-csv-consolidation`.
2. Go to `Permissions` -> `Grant Access`.
3. Principal: `wf21-scheduler-invoker@<PROJECT_ID>.iam.gserviceaccount.com`
4. Role: `Cloud Run Invoker`.
5. Save.

### 18.9 Create Cloud Scheduler Job In UI

1. Go to `Cloud Scheduler` -> `Create Job`.
2. Set:
   - Name: `wf21-drive-csv-every-5m`
   - Region: `asia-southeast1`
   - Frequency: `*/5 * * * *`
   - Time zone: `Asia/Singapore`
3. Target type: `HTTP`.
4. Method: `POST`.
5. URL:
   - `https://run.googleapis.com/v2/projects/<PROJECT_ID>/locations/asia-southeast1/jobs/wf21-drive-csv-consolidation:run`
6. Auth header:
   - Type: `OAuth token`
   - Service account: `wf21-scheduler-invoker@<PROJECT_ID>.iam.gserviceaccount.com`
   - Scope: `https://www.googleapis.com/auth/cloud-platform`
7. Click `Create`.
8. Test immediately: open scheduler job -> `Run now`.

### 18.10 UI Operations After Go-Live

1. Pause schedule:
   - `Cloud Scheduler` -> job -> `Pause`.
2. Resume schedule:
   - `Cloud Scheduler` -> job -> `Resume`.
3. Run on-demand:
   - `Cloud Run` -> Jobs -> job -> `Execute`.
4. Update image:
   - `Cloud Run` -> Jobs -> job -> `Edit & Deploy New Revision`.
5. Monitor:
   - `Cloud Run` execution logs
   - `Logs Explorer` with resource type `Cloud Run Job`.

## 19. Official References

- Cloud Run Jobs overview: https://cloud.google.com/run/docs/create-jobs
- Execute Jobs on schedule (Cloud Scheduler): https://cloud.google.com/run/docs/execute/jobs-on-schedule
- Cloud Run pricing: https://cloud.google.com/run/pricing
- Cloud Scheduler docs: https://cloud.google.com/scheduler/docs
- Secret Manager docs: https://cloud.google.com/secret-manager/docs
- Artifact Registry docs: https://cloud.google.com/artifact-registry/docs

