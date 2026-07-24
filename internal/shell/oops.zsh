# oops shell integration for zsh
# Installed by: oops --install

_oops_preexec() {
    export OOPS_LAST_CMD="$1"
    # Capture stderr via tee so the user still sees it in real-time.
    _OOPS_ERR=$(mktemp /tmp/.oops-XXXXXX 2>/dev/null) || return
    export _OOPS_ERR
    exec 3>&2 2> >(tee "$_OOPS_ERR" >&3 2>/dev/null)
}

_oops_precmd() {
    local _exit=$?
    # Restore stderr before anything else touches it.
    { exec 2>&3 3>&- ; } 2>/dev/null
    export OOPS_LAST_EXIT=$_exit
    if [[ $_exit -ne 0 && -n $_OOPS_ERR && -f $_OOPS_ERR ]]; then
        export OOPS_OUTPUT=$(<"$_OOPS_ERR")
    else
        export OOPS_OUTPUT=""
    fi
    [[ -n $_OOPS_ERR ]] && { rm -f "$_OOPS_ERR" 2>/dev/null; unset _OOPS_ERR; }
}

# Prepend precmd so we capture $? before any other hook modifies it.
precmd_functions=(_oops_precmd "${precmd_functions[@]}")
preexec_functions+=(_oops_preexec)

oops() {
    OOPS_HISTORY="$(fc -l -10 2>/dev/null)" \
    OOPS_CWD="$(pwd)" \
    OOPS_GIT_BRANCH="$(git branch --show-current 2>/dev/null)" \
    command oops "$@"
}
