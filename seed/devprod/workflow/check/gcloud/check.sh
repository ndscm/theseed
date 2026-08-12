#!/usr/bin/env bash
# Runner side of the ndscm check workflow (see report.sh for the reporting side).
#
# Migrated from seed/devprod/workflow/check/Jenkinsfile. These commands run the
# actual checks and therefore execute untrusted code from the commit under test
# (including its .envrc via direnv), so they are never given GH_TOKEN. Each phase
# records its outcome to the scratch dir for the paired report.sh step.
#
# Usage: check.sh prepare
#        check.sh run <phase>
#
# Environment (injected by cloudbuild.json / Cloud Build):
#   BRANCH_NAME   branch under test (prepare only)
set -euo pipefail

scratch_dir="/workspace/.check-ci"
testable_file="${scratch_dir}/testable"
failed_file="${scratch_dir}/failed"

clean_worktree() {
  git reset --hard HEAD
  git clean -fd
}

load_env() {
  cd /workspace
  direnv allow .
  eval "$(direnv export bash)"
}

prepare() {
  mkdir -p "${scratch_dir}"
  cd /workspace

  # Keep the scratch dir out of git's sight so it never trips the dirty check or
  # gets removed by clean_worktree.
  grep -qxF '/.check-ci/' .git/info/exclude 2>/dev/null ||
    echo '/.check-ci/' >>.git/info/exclude

  # Cloud Build may hand us a shallow clone without the branch ref that the
  # testable check needs; make sure the full history and the branch ref exist.
  git fetch --unshallow 2>/dev/null || true
  git fetch origin "+refs/heads/${BRANCH_NAME}:refs/remotes/origin/${BRANCH_NAME}"

  load_env

  local testable
  testable="$(ndscm testable --belong "refs/remotes/origin/${BRANCH_NAME}")"
  printf '%s' "${testable}" >"${testable_file}"
  echo "TESTABLE=${testable}"
}

# run executes one phase and records its outcome (pass|fail|skip) for the paired
# report.sh step. Format runs unconditionally; the rest are skipped on
# non-testable commits. A failure records a marker and still returns 0 so later
# phases run, matching the Jenkins catchError behaviour.
run() {
  local phase="$1"
  local label="${phase^}"
  local gate
  case "${phase}" in
  format) gate=always ;;
  *) gate=testable ;;
  esac

  load_env

  local rc=0
  if [[ "${gate}" == "testable" && "$(cat "${testable_file}")" != "true" ]]; then
    echo "Not testable commit, skipping ${label,,} phase."
    echo skip >"${scratch_dir}/result.${phase}"
  else
    (
      set -euo pipefail
      ndscm "${phase}"
      if [[ -n "$(git status --porcelain)" ]]; then
        echo "Worktree is dirty after ${label,,}:"
        git diff
        exit 1
      fi
    ) || rc=$?

    if [[ ${rc} -eq 0 ]]; then
      echo pass >"${scratch_dir}/result.${phase}"
    else
      echo fail >"${scratch_dir}/result.${phase}"
      touch "${failed_file}"
    fi
  fi

  clean_worktree
}

main() {
  local command="${1:-}"
  case "${command}" in
  prepare) prepare ;;
  run)
    shift
    run "$@"
    ;;
  *)
    echo "usage: check.sh prepare|run <phase>" >&2
    exit 2
    ;;
  esac
}

main "$@"
