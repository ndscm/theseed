function nd {
  if [[ "$#" -eq 0 ]]; then
    ndscm
  elif [[ "$#" -ge 1 ]]; then
    case "${1}" in
    "dev")
      eval "$(ndscm --shell-eval ${1} ${@:2})"
      ;;
    *)
      ndscm ${@:1}
      ;;
    esac
  fi
}
