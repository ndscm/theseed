#!/usr/bin/env bash
# Reporting side of the ndscm check workflow (see check.sh for the runner side).
#
# These commands only post GitHub commit statuses from outcomes recorded by
# check.sh; they never run repository code, so GH_TOKEN is confined to them.
#
# Usage: report.sh pending
#        report.sh report <phase>
#        report.sh finalize
#
# Environment (injected by cloudbuild.json / Cloud Build):
#   COMMIT_SHA        commit under test
#   CLOUD_BUILD_URL   console URL for this build, used as the status target_url
#   GH_TOKEN          GitHub token from Secret Manager (read by the gh CLI)
set -euo pipefail

repo="ndscm/theseed"
scratch_dir="/workspace/.check-ci"
failed_file="${scratch_dir}/failed"

phases=(format bootstrap tidy lock build test)

github_status() {
  local context="$1"
  local state="$2"
  local description="$3"
  gh api --method POST "/repos/${repo}/statuses/${COMMIT_SHA}" \
    -f context="${context}" \
    -f state="${state}" \
    -f description="${description}" \
    -f target_url="${CLOUD_BUILD_URL}"
}

pending() {
  github_status "ndscm/check" pending "Running check workflow"
  local phase
  for phase in "${phases[@]}"; do
    github_status "ndscm/${phase}" pending "Running ${phase}"
  done
}

report() {
  local phase="$1"
  local context="ndscm/${phase}"
  local label="${phase^}"

  local result
  result="$(cat "${scratch_dir}/result.${phase}")"

  case "${result}" in
  pass) github_status "${context}" success "${label} passed" ;;
  skip) github_status "${context}" success "${label} skipped" ;;
  fail) github_status "${context}" failure "${label} failed" ;;
  esac
}

finalize() {
  if [[ -f "${failed_file}" ]]; then
    github_status "ndscm/check" failure "Check workflow failed"
    exit 1
  fi
  github_status "ndscm/check" success "Check workflow passed"
}

main() {
  local command="${1:-}"
  case "${command}" in
  pending) pending ;;
  report)
    shift
    report "$@"
    ;;
  finalize) finalize ;;
  *)
    echo "usage: report.sh pending|report <phase>|finalize" >&2
    exit 2
    ;;
  esac
}

main "$@"
