# oops shell integration for zsh
# Installed by: oops --install

_oops_preexec() {
    export OOPS_LAST_CMD="$1"
}

_oops_precmd() {
    export OOPS_LAST_EXIT=$?
}

preexec_functions+=(_oops_preexec)
precmd_functions+=(_oops_precmd)

oops() {
    OOPS_HISTORY="$(fc -l -10 2>/dev/null)" \
    OOPS_CWD="$(pwd)" \
    OOPS_GIT_BRANCH="$(git branch --show-current 2>/dev/null)" \
    command oops "$@"
}
