# oops

Type `oops` after any failed command. The AI figures out what you meant and runs the fix.

```
$ git psuh origin main
git: 'psuh' is not a git command.

$ oops
oops: asking Claude (Anthropic)...

oops: detected a fix

  git push origin main

Execute? [Y/n/e(dit)] y

  $ git push origin main
Enumerating objects: 5, done.
```

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

**3. Set an LLM backend:**

```sh
export ANTHROPIC_API_KEY=sk-ant-...   # Claude haiku — fast, cheap, recommended
export OPENAI_API_KEY=sk-...          # GPT-4o-mini (or any OpenAI-compatible API)
# or install Ollama for fully local, private use
```

Add the export to your shell profile so it persists across sessions.

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

## LLM backends

Backends are checked in this order:

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

## How it works

oops hooks into your shell's preexec/precmd lifecycle to capture:

- The command that failed and its exit code
- Recent command history (for context about what you were trying to do)
- Working directory and current git branch

That context is sent to the LLM with a prompt that asks only for corrected shell commands — no explanations, no markdown. The response is parsed, safety-checked, and shown to you before anything runs.

## Supported shells

- zsh
- bash
- fish

## License

MIT
