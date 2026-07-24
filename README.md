# oops

Type `oops` after any failed command. It figures out what you meant and runs the fix — **no API key needed for most errors**.

```
$ git psuh origin main
git: 'psuh' is not a git command. See 'git --help'.

$ oops

oops: detected a fix

  git push origin main

Execute? [Y/n/e(dit)] y

  $ git push origin main
Enumerating objects: 5, done.
Counting objects: 100% (5/5), done.
Writing objects: 100% (3/3), 301 bytes, done.
```

A local rule engine handles the common stuff instantly. An LLM only gets involved when the rules don't have an answer.

## Install

### curl installer (recommended)

```sh
curl -fsSL https://raw.githubusercontent.com/karanlvm/oops/main/install.sh | sh
```

### go install

```sh
go install github.com/karanlvm/oops/cmd/oops@latest
```

### Build from source

```sh
git clone https://github.com/karanlvm/oops
cd oops
make install
```

## Setup

**1. Add shell hooks** (one-time):

```sh
oops --install
```

**2. Restart your shell**, or source your rc file:

```sh
source ~/.zshrc   # zsh
source ~/.bashrc  # bash
```

**3. (Optional) Set an LLM backend** for cases the local rules can't handle:

```sh
export ANTHROPIC_API_KEY=sk-ant-...   # Claude haiku — fast, cheap, recommended
export OPENAI_API_KEY=sk-...          # GPT-4o-mini (or any OpenAI-compatible endpoint)
# or just install Ollama — auto-detected, no key needed
```

Most common failures are handled without step 3. Add the export to your shell profile if you want LLM fallback to persist across sessions.

## Usage

```
oops              fix the last failed command
oops --yes        skip the confirmation prompt (safety checks still apply)
oops --dry-run    show the fix without executing it
oops --explain    explain why the fix works (coming soon)
oops --install    add shell hooks to your rc file
oops --uninstall  remove shell hooks
oops --version    print version
```

At the confirmation prompt, press `e` to open the suggested commands in `$EDITOR` before running.

## How it works

oops hooks into your shell's preexec/precmd lifecycle to capture the last failed command, its exit code, working directory, and git branch. Then it runs two layers:

**Layer 1 — local rule engine** (fires first, always)

A set of purpose-built rules that match common failure patterns and return a fix instantly. No network, no API key, no latency. Covers the failures you actually hit day-to-day.

**Layer 2 — LLM fallback** (only when Layer 1 finds nothing)

If no local rule matched, oops calls whichever LLM backend you've configured and sends the shell context as a prompt. The response is parsed, safety-checked, and shown to you before anything runs.

## Local rules coverage

| Rule | What it fixes |
|------|---------------|
| `git` | Subcommand typos (`psuh` → `push`, `chekcout` → `checkout`, and 100+ more), missing `-u` on first push, non-fast-forward push, missing `-m` on commit, branch `-d`/`-D` mismatch, rebase/merge conflict abort, and more |
| `typo` | 200+ direct substitutions for common shell command typos (`gti` → `git`, `claer` → `clear`, `pythoon` → `python3`, etc.) plus Levenshtein-1 fallback across 80+ known commands |
| `sudo` | Prepends `sudo` for commands that need root (package managers, systemctl, mount, etc.); `chmod +x` for scripts missing the execute bit |
| `cd` | Fuzzy-matches directory names so `cd porjects` finds `projects` in the current directory |
| `node` | npm/yarn/pnpm subcommand typos, reads `package.json` to fix `npm run <script-typo>`, suggests `npm install` before a run script |
| `make` | Reads `Makefile` targets and fuzzy-matches your typo to the closest real target |
| `python` | `python` → `python3`, `pip` → `pip3`, virtual environment activation hints |
| `docker` | Image/container typos, missing `docker compose up` before a run, `sudo` for permission errors |
| `brew` | Subcommand typos, `brew install` suggestions for unknown commands |
| `cargo` | Subcommand typos, `cargo build` before run when binary is missing |
| `go` | `go run`/`build`/`test` subcommand typos and common flag mistakes |

## LLM backends (optional enhancement)

If you set an API key, oops uses it as a fallback for anything the local rules don't cover. Backends are checked in this order:

| Backend | Env var | Model |
|---------|---------|-------|
| Anthropic | `ANTHROPIC_API_KEY` | claude-haiku-4-5 |
| OpenAI-compatible | `OPENAI_API_KEY` | gpt-4o-mini |
| Ollama | *(none)* | llama3.2 |

Ollama is auto-detected if it's running on `localhost:11434`. No key needed.

### OpenAI-compatible APIs (Groq, Together, etc.)

Point `OPENAI_BASE_URL` at any OpenAI-compatible endpoint:

```sh
export OPENAI_API_KEY=gsk_...
export OPENAI_BASE_URL=https://api.groq.com/openai/v1
```

## Safety

oops scans every suggested command for destructive patterns before running anything:

- `rm -rf`, `rm -fr`, `rm -r`
- `git reset --hard`, `git push --force`
- `git clean -f`
- `DROP TABLE`, `TRUNCATE TABLE`, `DELETE FROM`
- `kubectl delete`, `terraform destroy`
- pipe to shell (`| bash`, `| sh`, etc.)

If a match is found, oops prints a warning and requires you to type `YES` (all caps) to proceed. The `--yes` flag does **not** bypass this — it only skips the normal Y/n prompt.

## Supported shells

- zsh
- bash
- fish

## License

MIT
