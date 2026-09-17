# Deploying the onboarding-nudges Cloud Run Jobs

Two jobs, same pattern as `cmd/enquiry-autoreject` (mirrors its
`Dockerfile`/`cloudbuild.yaml` exactly):

- **`onboarding-nudges`** — the recurring sweep. Deploy once, then put it on
  a Cloud Scheduler cadence (every 15–30 min) and forget it.
- **`onboarding-nudge-backfill`** — a one-time tool. Deploy once, run it
  once with `-apply`, done. Not scheduled.

All commands assume you're in the `mandi-backend/` repo root, authenticated
(`gcloud auth login`) against project `locazar-f20b6`, region `us-central2`
(matching the substitutions already in both `cloudbuild.yaml` files).

```bash
export PROJECT=locazar-f20b6
export REGION=us-central2
gcloud config set project "$PROJECT"
```

## 0. One-time project setup (skip anything already enabled)

```bash
gcloud services enable run.googleapis.com cloudbuild.googleapis.com \
  cloudscheduler.googleapis.com secretmanager.googleapis.com \
  --project="$PROJECT"
```

### Config: `.env` for the job

**This is the part that will bite you if skipped.** `config.LoadConfig()`
(`pkg/config/config.go`) hard-requires a `.env` file in the working
directory at startup (`viper.SetConfigFile(".env")`) — it does not fall back
to real process environment variables. Neither this job's `Dockerfile` nor
`cmd/enquiry-autoreject`'s copies one in (on purpose — a `.env` full of DB
credentials and Firebase keys doesn't belong baked into an image layer).
Without one mounted at runtime, the job crashes immediately on `LoadConfig`.

Store it in Secret Manager once, then mount it into every job that needs it:

```bash
# One-time: upload your real .env (same one the API server uses)
gcloud secrets create mandi-backend-env --data-file=.env --project="$PROJECT"
# Later, when secrets rotate: gcloud secrets versions add mandi-backend-env --data-file=.env
```

Every `gcloud run jobs create`/`update` below mounts it at `/app/.env` via
`--set-secrets`, matching the working directory (`WORKDIR /` in the
Dockerfile, so `.env` resolves as `/.env` — adjust the mount path below to
`/.env` if `LoadConfig` can't find it; test with `--execute-now` and check
logs before wiring the scheduler).

If `enquiry-autoreject` is already deployed and working, check how *it*
supplies `.env` instead of guessing — it's the same problem, solved once:

```bash
gcloud run jobs describe enquiry-autoreject --region="$REGION" --project="$PROJECT" \
  --format="yaml(spec.template.spec.template.spec.containers[0].env, spec.template.spec.template.spec.volumes)"
```

If that command 404s, `enquiry-autoreject` was never actually deployed
either — treat both jobs as new infrastructure.

## 1. onboarding-nudges — the recurring sweep

### Build & push the image

```bash
gcloud builds submit . \
  --project="$PROJECT" \
  --config=cmd/onboarding-nudges/cloudbuild.yaml \
  --substitutions=COMMIT_SHA=v1,_REGION="$REGION",_JOB_NAME=onboarding-nudges
```

This builds `cmd/onboarding-nudges/Dockerfile`, pushes
`gcr.io/$PROJECT/onboarding-nudges:v1`, then tries `gcloud run jobs update`
— which **fails the first time** because the job doesn't exist yet. That's
expected; the push still succeeds. Create the job once:

```bash
gcloud run jobs create onboarding-nudges \
  --image="gcr.io/$PROJECT/onboarding-nudges:v1" \
  --region="$REGION" \
  --project="$PROJECT" \
  --set-secrets="/app/.env=mandi-backend-env:latest" \
  --max-retries=1 \
  --task-timeout=600
```

From now on, re-running the `gcloud builds submit` command above (bump
`COMMIT_SHA` each time, e.g. `v2`, `v3`, or wire a real Cloud Build trigger
on push) both rebuilds and updates the job in one step.

### Verify it runs before scheduling it

```bash
gcloud run jobs execute onboarding-nudges --region="$REGION" --project="$PROJECT" --wait
gcloud logging read \
  'resource.type=cloud_run_job AND resource.labels.job_name=onboarding-nudges' \
  --project="$PROJECT" --limit=50 --order=desc
```

Look for the `onboarding-nudges: sent=X skipped=Y errors=Z` line the binary
logs on exit (see `cmd/onboarding-nudges/main.go`). `errors=0` and no crash
= good.

### Wire Cloud Scheduler (every 20 minutes)

Cloud Scheduler invokes the Cloud Run Jobs Admin API directly over HTTPS,
authenticated via a service account's OIDC token — no shared secret needed.

```bash
# One-time: a service account with just enough permission to run this one job
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

`*/20 * * * *` = every 20 minutes, inside your 15–30 min target. Adjust
freely — the sweep is idempotent and catches up on any gap (see the doc
comment in `cmd/onboarding-nudges/main.go`), so cadence is a tuning knob,
not a correctness concern.

## 2. onboarding-nudge-backfill — run once, after the above is live

```bash
gcloud builds submit . \
  --project="$PROJECT" \
  --config=cmd/onboarding-nudge-backfill/cloudbuild.yaml \
  --substitutions=COMMIT_SHA=v1,_REGION="$REGION",_JOB_NAME=onboarding-nudge-backfill

gcloud run jobs create onboarding-nudge-backfill \
  --image="gcr.io/$PROJECT/onboarding-nudge-backfill:v1" \
  --region="$REGION" \
  --project="$PROJECT" \
  --set-secrets="/app/.env=mandi-backend-env:latest" \
  --max-retries=0 \
  --task-timeout=600

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
*before* running this with `-apply`, since it can't be undone (only future
sends can be turned off, via the Enabled toggle).

## 3. Don't want to wait for the scheduler?

admin-portal's Onboarding Nudges page has a **"Run sweep now"** button
(`POST /api/admin/onboarding-nudges/run-sweep`, real JWT admin auth, same
sweep logic) — use it to trigger an on-demand run any time, independent of
whether the Cloud Scheduler job above is set up yet.
