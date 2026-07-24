# oops shell integration for bash
# Installed by: oops --install

_oops_preexec() {
    export OOPS_LAST_CMD="$BASH_COMMAND"
}

_oops_precmd() {
    export OOPS_LAST_EXIT=$?
}

trap '_oops_preexec' DEBUG

PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND; }_oops_precmd"

oops() {
    OOPS_HISTORY="$(history 10 2>/dev/null)" \
    OOPS_CWD="$(pwd)" \
    OOPS_GIT_BRANCH="$(git branch --show-current 2>/dev/null)" \
    command oops "$@"
}
