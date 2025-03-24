if [[ -z "${ND_MONOREPO_HOME+x}" &&
  -d "${ND_MONOREPO_HOME}/" &&
  -z "${ND_MONOREPO_GIT_DIR+x}" &&
  -d "${ND_MONOREPO_GIT_DIR}/" ]]; then
  export ND_MONOREPO_HOME=$(realpath "${ND_MONOREPO_HOME}/")
  export ND_MONOREPO_GIT_DIR=$(realpath "${ND_MONOREPO_GIT_DIR}/")
fi

if [[ -z "${NDSCM_HOME+x}" ]]; then
  # Guess NDSCM_HOME from source file location
  if [[ ! -z "${BASH_SOURCE+x}" ]]; then
    script_location="${BASH_SOURCE[0]}"
  else
    script_location="${0}"
  fi
  export NDSCM_HOME=$(realpath $(dirname "${script_location}"))
fi
if [[ ! -d "${NDSCM_HOME}" || ! -f "${NDSCM_HOME}/envsetup.sh" ]]; then
  echo -e "\e[31mNDSCM_HOME (${NDSCM_HOME}) is not valid\e[0m"
  return 1
fi

function nd-quick-verify-monorepo {
  if [[ -z "${ND_MONOREPO_HOME+x}" ]]; then
    echo -e "\e[31mND_MONOREPO_HOME is not set\e[0m"
    return 1
  fi
  if [[ -z "${ND_MONOREPO_GIT_DIR+x}" ]]; then
    echo -e "\e[31mND_MONOREPO_GIT_DIR is not set\e[0m"
    return 1
  fi
  if [[ ! -d "${ND_MONOREPO_HOME}" ]]; then
    echo -e "\e[31mND_MONOREPO_HOME (${ND_MONOREPO_HOME}) is not a valid folder\e[0m"
    return 1
  fi
  if [[ ! -d "${ND_MONOREPO_GIT_DIR}" ]]; then
    echo -e "\e[31mND_MONOREPO_GIT_DIR (${ND_MONOREPO_GIT_DIR}) is not a valid folder\e[0m"
    return 1
  fi

}

function nd-dev {
  if ! nd-quick-verify-monorepo; then
    return $?
  fi
  if [[ "$#" -eq 0 ]]; then
    cd "${ND_MONOREPO_HOME}/dev"
  elif [[ "$#" -eq 1 ]]; then
    worktree_name="dev-${1}"
    if [[ -d "${ND_MONOREPO_HOME}/${worktree_name}" ]]; then
      cd "${ND_MONOREPO_HOME}/${worktree_name}"
    else
      git --git-dir "${ND_MONOREPO_GIT_DIR}" worktree add -B "${worktree_name}" "${ND_MONOREPO_HOME}/${worktree_name}" origin/main
      cd "${ND_MONOREPO_HOME}/${worktree_name}"
    fi
  else
    echo -e "\e[31mToo many params for nd-dev.\e[0m"
    return 1
  fi
}

function nd {
  if [[ "$#" -eq 0 ]]; then
    echo "nd is not distributed source code manager"
  elif [[ "$#" -ge 1 ]]; then
    case "${1}" in
    "dev")
      nd-dev ${@:2}
      ;;
    *)
      echo -e "\e[31mUnknown command ${1}\e[0m"
      ;;
    esac
  else
    echo -e "\e[31mUnknown params ${@}\e[0m"
    return 1
  fi
}
