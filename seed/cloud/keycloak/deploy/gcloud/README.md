# Keycloak Deploy Script (Google Cloud Run)

Script to deploy Keycloak to Cloud Run.

## Deploy

```bash
# From the monorepo root
./seed/cloud/keycloak/deploy/gcloud/deploy.sh
```

This builds the container image, pushes it to Artifact Registry, and updates the
Cloud Run service.

## One-time setup

The deploy script runs `gcloud run services update` (not `deploy`), so the Cloud
Run service needs to already exist with:

- `KC_HOSTNAME` env set to the public hostname (e.g. `account.ndscm.com`)
- Secret mounts (e.g. `keycloak-database-secret-prod` for `KC_DB_PASSWORD`)
- `--add-cloudsql-instances ndscm-prod:us-west1:ndscm-cloudsql-prod`
- Service account `keycloak-prod@ndscm-prod.iam.gserviceaccount.com` with the
  Cloud SQL Client role

## Notes

**Local image cache dependency:** The deploy Dockerfile pulls
`FROM ghcr.io/ndscm/seed-cloud-keycloak-container:latest` out of the local
Docker cache (the container build doesn't push to a registry). This is fine for
developer-driven deploys but will break on a clean CI runner.

**`--no-cpu-throttling`:** Required for services like Keycloak that respond to
requests immediately and trigger background jobs. Without this, CPU is throttled
outside of request handling, which would kill background work.
