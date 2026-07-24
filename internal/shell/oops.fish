# oops shell integration for fish
# Installed by: oops --install

function _oops_preexec --on-event fish_preexec
    set -gx OOPS_LAST_CMD $argv[1]
end

function _oops_precmd --on-event fish_postexec
    set -gx OOPS_LAST_EXIT $status
end

function oops
    set -lx OOPS_HISTORY (builtin history | head -10 | string join "\n")
    set -lx OOPS_CWD (pwd)
    set -lx OOPS_GIT_BRANCH (git branch --show-current 2>/dev/null)
    command oops $argv
end
