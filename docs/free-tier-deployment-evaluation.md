# WF21 Deployment Fit (Free Tier) - March 12, 2026

## Goal
Assess if this workflow is a good fit for these free-tier platforms and recommend a stable deployment setup (with UptimeRobot pinging).

Workflow shape:
- Long-running Go worker (`WF21_CONTINUOUS=true`)
- Polls Google Drive every second
- Writes to Google Sheets + Cloudflare R2
- Exposes health endpoint (`/healthz`) for uptime monitoring

## Quick Verdict
Best fit on free tier: **Render + GitHub + Cloudflare (R2) + UptimeRobot**.

Reason:
- This workflow is a **continuous worker**, not a short serverless function.
- Most serverless/free options here have runtime/scheduling limits that are not ideal for 24/7 polling.

## Platform-by-Platform Fit

| Platform | Fit for this workflow | Why |
|---|---|---|
| Render (no blueprint) | **Good (best among listed free tiers)** | Supports long-running web service. Free service spins down after 15m idle and has 750 free instance hours/month, so monitor pinging matters. |
| Vercel | **Poor** | Hobby cron is restricted (daily), and cron/function model is not a good fit for a continuous poller. |
| Railway | **Possible but fragile on free** | Free plan is credit-based (`$1/month` after trial), so uptime depends on usage cost. |
| Supabase | **Poor for hosting this worker** | Edge Functions on free have short duration limits; free projects can pause for inactivity. Better as data backend, not as always-on worker host here. |
| Cloudflare (Workers) | **Poor without redesign** | Free Workers have strict CPU/request limits; would require redesign to event/cron architecture. |
| GitHub Actions | **Poor for primary runtime** | Scheduled workflows are not designed for near-real-time polling loops; schedule has minimum interval and can be delayed. |

## Recommended Free-Tier Stack

1. **Compute runtime**: Render Free Web Service (single instance).
2. **Code hosting + CI/CD**: GitHub.
3. **Object/state storage**: Cloudflare R2 (already used by workflow).
4. **External liveness monitor**: UptimeRobot (free, 5-minute checks).

## Stability Hardening Checklist

1. Keep exactly one runtime instance:
   - Set and keep `WF21_LOCK_FILE` configured.
   - Never run a second manual `go run ./cmd` while Render service is active.
2. Use persistent state outside local disk:
   - On Render free, filesystem is ephemeral.
   - Prefer `WF21_STATE_FILE`, `WF21_STATUS_FILE`, and lock file on `r2://...` keys to survive restarts/redeploys.
3. Keep health endpoint enabled:
   - `WF21_ENABLE_HEALTH_SERVER=true`
   - UptimeRobot monitor: `GET /healthz`
4. Keep summary send robust:
   - Current flow is: import 3 tabs -> update `config!B1` -> wait 8s -> wait for stable image sources -> send.
5. Keep retries and warnings visible:
   - Review logs for `post-import warning ...` lines.

## Suggested Deployment Choice (Now)

Use **Render** now for fastest stable deployment with your current codebase.

If you later want stronger reliability than free-tier constraints:
- Move to paid always-on compute (Render paid instance or Railway paid).
- Keep state/lock/status in R2 instead of local files.

## Source Links

- Render free behavior and limits (spin-down + 750 free instance hours): https://render.com/free
- Render filesystem and persistent disk constraints (ephemeral by default; disks on paid services): https://render.com/docs/disks
- Vercel cron pricing/limits (Hobby restrictions): https://vercel.com/docs/cron-jobs/usage-and-pricing
- Vercel cron management notes (timing accuracy + function duration linkage): https://vercel.com/docs/cron-jobs/manage-cron-jobs
- Railway free trial and free plan credit model: https://docs.railway.com/reference/pricing/free-trial
- Railway plan reference: https://docs.railway.com/reference/pricing/plans
- Supabase Edge Function limits (free duration): https://supabase.com/docs/guides/functions/limits
- Supabase free inactivity pause note: https://supabase.com/docs/guides/deployment/going-into-prod
- Cloudflare Workers limits (free requests/CPU/cron trigger limits): https://developers.cloudflare.com/workers/platform/limits/
- GitHub Actions schedule interval and behavior (5-minute minimum): https://docs.github.com/enterprise-cloud%40latest/actions/using-workflows/workflow-syntax-for-github-actions
- GitHub scheduled workflow delay note: https://docs.github.com/en/actions/concepts/workflows-and-actions/about-troubleshooting-workflows
- UptimeRobot free monitoring interval (5 minutes): https://help.uptimerobot.com/en/articles/11360876-what-is-a-monitoring-interval-in-uptimerobot
