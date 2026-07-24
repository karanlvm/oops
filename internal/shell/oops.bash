# oops shell integration for bash
# Installed by: oops --install

_oops_fired=0

_oops_preexec() {
    [[ $_oops_fired -eq 1 ]] && return
    _oops_fired=1
    export OOPS_LAST_CMD="$BASH_COMMAND"
    _OOPS_ERR=$(mktemp /tmp/.oops-XXXXXX 2>/dev/null) || return
    export _OOPS_ERR
    exec 3>&2 2> >(tee "$_OOPS_ERR" >&3 2>/dev/null)
}

_oops_precmd() {
    local _exit=$?
    _oops_fired=0
    { exec 2>&3 3>&- ; } 2>/dev/null
    export OOPS_LAST_EXIT=$_exit
    if [[ $_exit -ne 0 && -n $_OOPS_ERR && -f $_OOPS_ERR ]]; then
        export OOPS_OUTPUT=$(<"$_OOPS_ERR")
    else
        export OOPS_OUTPUT=""
    fi
    [[ -n $_OOPS_ERR ]] && { rm -f "$_OOPS_ERR" 2>/dev/null; unset _OOPS_ERR; }
}

trap '_oops_preexec' DEBUG
PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND; }_oops_precmd"

oops() {
    OOPS_HISTORY="$(history 10 2>/dev/null)" \
    OOPS_CWD="$(pwd)" \
    OOPS_GIT_BRANCH="$(git branch --show-current 2>/dev/null)" \
    command oops "$@"
}
