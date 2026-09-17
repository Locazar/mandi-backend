# Deploying the onboarding-nudges Cloud Run Jobs

**Superseded for the recurring sweep.** `cmd/api/main.go` now runs the sweep
itself, in-process, every 15 minutes, for as long as the API server is up —
see `runOnboardingNudgeTicker`. This works today with zero extra setup: the
API server already has real DB credentials, and
`deploy/ecommerce-deployment.yaml` runs it as a single replica, so there's
no multi-replica double-fire risk a ticker would otherwise have. "Start" and
"stop" are the existing Enabled toggle on admin-portal's Onboarding Nudges
page — the ticker always runs, but `RunSweep` no-ops on every tick when
Enabled is off, so there's nothing separate to start/stop.

Everything below (the standalone Cloud Run Job) is now optional — a fallback
if the ticker ever needs to be decoupled from the API server's own uptime.
`onboarding-nudge-backfill` (the one-time tool, not the recurring sweep)
still uses this Cloud Run Job path either way, since it's a single manual
run, not something that belongs ticking inside the API server.

Two jobs:

- **`onboarding-nudges`** — the recurring sweep. Once credentials are sorted,
  put it on a Cloud Scheduler cadence (every 15–30 min) and forget it.
- **`onboarding-nudge-backfill`** — a one-time tool. Run it once with
  `-apply` after credentials are sorted, done. Not scheduled.

Region is **`asia-south1`** — not `us-central2`, which this project's org
policy rejects (`LOCATION_POLICY_VIOLATED`). Images live in a new Artifact
Registry repo, `mandi-backend` (asia-south1) — this project had no
general-purpose registry before this; the only pre-existing repos are
Cloud Functions' auto-managed `gcf-artifacts`, not meant for manual pushes.

```bash
export PROJECT=locazar-f20b6
export REGION=asia-south1
gcloud config set project "$PROJECT"
```

## The credentials gap (read this first)

`config.LoadConfig()` (`pkg/config/config.go:234-257`) reads a `.env` file
if present, but the error from a missing one is explicitly discarded
(`_ = viper.ReadInConfig()`) — so it does NOT crash without one. Instead it
falls back to real process env vars via `viper.BindEnv(...)` for each
config key. So a job needs *either* a mounted `.env` *or* every required
var (`DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `FIREBASE_*`, etc. —
see `pkg/config/config.go` for the full list) set directly as container env
vars.

**Neither exists yet for these two new jobs.** I checked whether
`cmd/enquiry-autoreject` (already deployed, as a Cloud Run *Service* in
`asia-south1` — its own doc comment claiming "Cloud Run Job" doesn't match
reality) could be copied as a working example: it has exactly two env vars
(`ENQUIRY_AUTO_REJECT_HOURS`, `LOG_EXECUTION_ID`), no secrets, no volume
mounts, nothing DB/Firebase-related at all. That means it cannot actually
reach the database as currently deployed either — it's not a working
reference, so there was nothing safe to copy.

No local `.env` exists in this repo (correctly — it shouldn't), and no
secret named anything like `mandi-backend-env` exists in Secret Manager
(`gcloud secrets list` returns empty). **I don't have the real DB/Firebase
credentials and won't fabricate placeholder ones into a "real" resource** —
wrong values could point at the wrong database, which is worse than doing
nothing.

**To unblock, you need to supply one of:**

1. The real `.env` this project's API server actually uses, so I can:
   ```bash
   gcloud secrets create mandi-backend-env --data-file=/path/to/real/.env --project="$PROJECT"
   gcloud run jobs update onboarding-nudges --region="$REGION" --project="$PROJECT" \
     --set-secrets="/app/.env=mandi-backend-env:latest"
   gcloud run jobs update onboarding-nudge-backfill --region="$REGION" --project="$PROJECT" \
     --set-secrets="/app/.env=mandi-backend-env:latest"
   ```
2. Or tell me where the *real* production database config actually lives
   today (a different secret name already in this project? a different GCP
   project? the AWS/K8s side entirely, per `deploy/ecommerce-deployment.yaml`'s
   `postgres-secret`?) — if it's the latter, these DB credentials may not be
   reachable from a GCP Cloud Run Job at all without a Cloud SQL proxy /
   VPC connector, which changes this setup meaningfully.

## What's already done

```bash
# Images built and pushed:
asia-south1-docker.pkg.dev/locazar-f20b6/mandi-backend/onboarding-nudges:v1
asia-south1-docker.pkg.dev/locazar-f20b6/mandi-backend/onboarding-nudge-backfill:v1

# Jobs created (inert — no secrets, never executed):
gcloud run jobs describe onboarding-nudges --region=asia-south1 --project=locazar-f20b6
gcloud run jobs describe onboarding-nudge-backfill --region=asia-south1 --project=locazar-f20b6
```

Rebuilding after a code change (bump the tag each time, e.g. `v2`):

```bash
gcloud builds submit . --project="$PROJECT" \
  --config=cmd/onboarding-nudges/cloudbuild.yaml \
  --substitutions=COMMIT_SHA=v2,_REGION="$REGION",_JOB_NAME=onboarding-nudges
```

This rebuilds, pushes, and runs `gcloud run jobs update` in one step — that
update step only works once the job exists (see "What's already done"
above); the very first build for a new job will report that step as failed,
which is expected — the image still pushed successfully.

## Once credentials are wired: verify before scheduling

```bash
gcloud run jobs execute onboarding-nudges --region="$REGION" --project="$PROJECT" --wait
gcloud logging read \
  'resource.type=cloud_run_job AND resource.labels.job_name=onboarding-nudges' \
  --project="$PROJECT" --limit=50 --order=desc
```

Look for `onboarding-nudges: sent=X skipped=Y errors=Z` (see
`cmd/onboarding-nudges/main.go`). `errors=0` and no crash = good.

## Wire Cloud Scheduler (every 20 minutes)

Cloud Scheduler invokes the Cloud Run Jobs Admin API directly over HTTPS,
authenticated via a service account's OIDC token — no shared secret needed.

```bash
gcloud iam service-accounts create onboarding-nudges-invoker \
  --display-name="Invokes the onboarding-nudges Cloud Run Job" \
  --project="$PROJECT"

gcloud run jobs add-iam-policy-binding onboarding-nudges \
  --region="$REGION" --project="$PROJECT" \
  --member="serviceAccount:onboarding-nudges-invoker@${PROJECT}.iam.gserviceaccount.com" \
  --role="roles/run.invoker"

gcloud scheduler jobs create http onboarding-nudges-sweep \
  --project="$PROJECT" \
  --location="$REGION" \
  --schedule="*/20 * * * *" \
  --uri="https://${REGION}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${PROJECT}/jobs/onboarding-nudges:run" \
  --http-method=POST \
  --oauth-service-account-email="onboarding-nudges-invoker@${PROJECT}.iam.gserviceaccount.com"
```

`*/20 * * * *` = every 20 minutes, inside the 15–30 min target. Cadence is a
tuning knob, not a correctness concern — the sweep is idempotent and catches
up on any gap (see `cmd/onboarding-nudges/main.go`'s doc comment).

## onboarding-nudge-backfill — run once, after everything above works

```bash
# Dry-run first — prints what it WOULD anchor, writes nothing:
gcloud run jobs execute onboarding-nudge-backfill --region="$REGION" --project="$PROJECT" --wait
gcloud logging read \
  'resource.type=cloud_run_job AND resource.labels.job_name=onboarding-nudge-backfill' \
  --project="$PROJECT" --limit=100 --order=desc

# Looks right? Actually write the anchors:
gcloud run jobs execute onboarding-nudge-backfill --region="$REGION" --project="$PROJECT" \
  --args="-apply" --wait
```

**Before running `-apply`**: every currently-active shop gets anchored at
the same "now" timestamp, so they'll all become due for their first nudge
(Add Products) together on the very next `onboarding-nudges` sweep — a
synchronized burst across every existing seller, not a trickle. Make sure
the schedule/templates in admin-portal's Onboarding Nudges page look right
*before* running this with `-apply` — it can't be undone (only future sends
can be turned off, via the Enabled toggle).

## Don't want to wait for the scheduler?

admin-portal's Onboarding Nudges page has a **"Run sweep now"** button
(`POST /api/admin/onboarding-nudges/run-sweep`, real JWT admin auth, same
sweep logic) — use it to trigger an on-demand run any time. It calls the
*live API server*, not the Cloud Run Job, so it only works once the API
server itself has been redeployed with this code — separate from everything
above, and a pipeline I don't have visibility into (see the main handoff
message for why).
