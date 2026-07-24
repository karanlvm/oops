package rules_test

import (
	"os"
	"path/filepath"
	"testing"

	ctx "github.com/karanlvm/oops/internal/context"
	"github.com/karanlvm/oops/internal/rules"
)

func suggest(cmd string, exit int) []string {
	return rules.Suggest(cmd, exit, ctx.ShellContext{
		LastCommand:  cmd,
		LastExitCode: exit,
	})
}

func suggestCtx(cmd string, exit int, c ctx.ShellContext) []string {
	c.LastCommand = cmd
	c.LastExitCode = exit
	return rules.Suggest(cmd, exit, c)
}

// ── git ──────────────────────────────────────────────────────────────────────

func TestGitTypos(t *testing.T) {
	cases := []struct {
		cmd  string
		want string
	}{
		{"git psuh origin main", "git push origin main"},
		{"git pish origin main", "git push origin main"},
		{"git comit -m 'fix'", "git commit -m 'fix'"},
		{"git comitt -m 'fix'", "git commit -m 'fix'"},
		{"git chekcout main", "git checkout main"},
		{"git statsu", "git status"},
		{"git sttaus", "git status"},
		{"git brnach -a", "git branch -a"},
		{"git branh -D feat", "git branch -D feat"},
		{"git reabse main", "git rebase main"},
		{"git stsh pop", "git stash pop"},
		{"git feth", "git fetch"},
		{"git dfif HEAD", "git diff HEAD"},
		{"git lgo --oneline", "git log --oneline"},
		{"git meger main", "git merge main"},
		{"git cloen https://github.com/x/y", "git clone https://github.com/x/y"},
		{"git addd .", "git add ."},
		{"git swich main", "git switch main"},
		{"git shwo HEAD", "git show HEAD"},
		{"git resot HEAD~1", "git reset HEAD~1"},
		{"git remtoe -v", "git remote -v"},
		{"git tga v1.0", "git tag v1.0"},
		{"gti status", "git status"},
		{"gti push", "git push"},
	}
	for _, tc := range cases {
		got := suggest(tc.cmd, 1)
		if len(got) == 0 || got[0] != tc.want {
			t.Errorf("Suggest(%q): got %v, want [%q]", tc.cmd, got, tc.want)
		}
	}
}

func TestGitLevenshteinFallback(t *testing.T) {
	// Words not in the typo table but within edit distance 2.
	cases := []struct{ cmd, want string }{
		{"git stat", "git stash"},   // dist 2 — closest is stash
		{"git tgas v1.0", "git tag v1.0"}, // dist 1
	}
	for _, tc := range cases {
		got := suggest(tc.cmd, 1)
		if len(got) == 0 {
			t.Errorf("Suggest(%q): got no fix, want %q", tc.cmd, tc.want)
		}
	}
}

func TestGitNoUpstream(t *testing.T) {
	c := ctx.ShellContext{GitBranch: "feat/foo"}
	got := suggestCtx("git push", 128, c)
	if len(got) == 0 || got[0] != "git push -u origin feat/foo" {
		t.Errorf("got %v", got)
	}
}

func TestGitPullNoUpstream(t *testing.T) {
	c := ctx.ShellContext{GitBranch: "main"}
	got := suggestCtx("git pull", 128, c)
	if len(got) == 0 || got[0] != "git pull origin main" {
		t.Errorf("got %v", got)
	}
}

func TestGitCommitMissingFlag(t *testing.T) {
	got := suggest(`git commit "my message"`, 1)
	if len(got) == 0 || got[0] != `git commit -m "my message"` {
		t.Errorf("got %v", got)
	}
}

func TestGitBranchDeleteUnmerged(t *testing.T) {
	got := suggest("git branch -d feat/old", 1)
	if len(got) == 0 || got[0] != "git branch -D feat/old" {
		t.Errorf("got %v", got)
	}
}

func TestGitCheckoutExistingBranch(t *testing.T) {
	got := suggest("git checkout -b main", 128)
	if len(got) == 0 || got[0] != "git checkout main" {
		t.Errorf("got %v", got)
	}
}

func TestGitCleanNeedsForce(t *testing.T) {
	got := suggest("git clean -d", 1)
	if len(got) == 0 {
		t.Errorf("expected fix for git clean without -f")
	}
}

// ── sudo ─────────────────────────────────────────────────────────────────────

func TestSudoPermissionDenied(t *testing.T) {
	got := suggest("apt install curl", 1)
	if len(got) == 0 || got[0] != "sudo apt install curl" {
		t.Errorf("got %v", got)
	}
}

func TestSudoExit126(t *testing.T) {
	got := suggest("somebinary --flag", 126)
	if len(got) == 0 || got[0] != "sudo somebinary --flag" {
		t.Errorf("got %v", got)
	}
}

func TestSudoScriptChmod(t *testing.T) {
	got := suggest("./deploy.sh --prod", 126)
	if len(got) < 2 || got[0] != "chmod +x ./deploy.sh" || got[1] != "./deploy.sh --prod" {
		t.Errorf("got %v", got)
	}
}

func TestSudoAlreadySudo(t *testing.T) {
	got := suggest("sudo apt install curl", 1)
	if len(got) != 0 {
		t.Errorf("should not suggest sudo when already present: got %v", got)
	}
}

// ── cd ───────────────────────────────────────────────────────────────────────

func TestCdTypo(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "myproject")
	os.Mkdir(target, 0755)

	c := ctx.ShellContext{WorkingDir: dir}
	got := suggestCtx("cd myproejct", 1, c)
	if len(got) == 0 || got[0] != "cd myproject" {
		t.Errorf("got %v", got)
	}
}

func TestCdNoMatch(t *testing.T) {
	dir := t.TempDir()
	c := ctx.ShellContext{WorkingDir: dir}
	got := suggestCtx("cd totallynonexistentxyz", 1, c)
	if len(got) != 0 {
		t.Errorf("expected no fix for unrecognisable dir, got %v", got)
	}
}

// ── python ───────────────────────────────────────────────────────────────────

func TestPythonBinaryTypo(t *testing.T) {
	cases := []struct{ cmd, want string }{
		{"pyhton script.py", "python3 script.py"},
		{"pytohn -c 'print(1)'", "python3 -c 'print(1)'"},
		{"python script.py", "python3 script.py"},
		{"pip install numpy", "pip3 install numpy"},
	}
	for _, tc := range cases {
		got := suggest(tc.cmd, 127)
		if len(got) == 0 || got[0] != tc.want {
			t.Errorf("Suggest(%q): got %v, want [%q]", tc.cmd, got, tc.want)
		}
	}
}

// ── node ─────────────────────────────────────────────────────────────────────

func TestNpmSubcmdTypo(t *testing.T) {
	got := suggest("npm insatll", 1)
	if len(got) == 0 || got[0] != "npm install" {
		t.Errorf("got %v", got)
	}
}

func TestNpmRunScriptTypo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"scripts": {"build": "tsc", "test": "jest", "dev": "next dev"}
	}`), 0644)

	c := ctx.ShellContext{WorkingDir: dir}
	got := suggestCtx("npm run buidl", 1, c)
	if len(got) == 0 || got[0] != "npm run build" {
		t.Errorf("got %v", got)
	}
}

func TestYarnScriptTypo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{
		"scripts": {"build": "tsc", "start": "node index.js"}
	}`), 0644)

	c := ctx.ShellContext{WorkingDir: dir}
	got := suggestCtx("yarn biuld", 1, c)
	if len(got) == 0 || got[0] != "yarn build" {
		t.Errorf("got %v", got)
	}
}

// ── docker ───────────────────────────────────────────────────────────────────

func TestDockerCompose(t *testing.T) {
	got := suggest("docker-compose up -d", 1)
	if len(got) == 0 || got[0] != "docker compose up -d" {
		t.Errorf("got %v", got)
	}
}

func TestDockerBinaryTypo(t *testing.T) {
	got := suggest("dokcer ps", 127)
	if len(got) == 0 || got[0] != "docker ps" {
		t.Errorf("got %v", got)
	}
}

func TestDockerSubcmdTypo(t *testing.T) {
	got := suggest("docker pss", 1)
	if len(got) == 0 || got[0] != "docker ps" {
		t.Errorf("got %v", got)
	}
}

// ── brew ─────────────────────────────────────────────────────────────────────

func TestBrewBinaryTypo(t *testing.T) {
	got := suggest("brwe install git", 127)
	if len(got) == 0 || got[0] != "brew install git" {
		t.Errorf("got %v", got)
	}
}

func TestBrewSubcmdTypo(t *testing.T) {
	got := suggest("brew insatll git", 1)
	if len(got) == 0 || got[0] != "brew install git" {
		t.Errorf("got %v", got)
	}
}

// ── cargo ─────────────────────────────────────────────────────────────────────

func TestCargoBinaryTypo(t *testing.T) {
	got := suggest("craog build", 127)
	if len(got) == 0 || got[0] != "cargo build" {
		t.Errorf("got %v", got)
	}
}

func TestCargoSubcmdTypo(t *testing.T) {
	got := suggest("cargo biuld", 1)
	if len(got) == 0 || got[0] != "cargo build" {
		t.Errorf("got %v", got)
	}
}

// ── go ───────────────────────────────────────────────────────────────────────

func TestGoSubcmdTypo(t *testing.T) {
	cases := []struct{ cmd, want string }{
		{"go buidl ./...", "go build ./..."},
		{"go tset ./...", "go test ./..."},
		{"go rnu main.go", "go run main.go"},
	}
	for _, tc := range cases {
		got := suggest(tc.cmd, 1)
		if len(got) == 0 || got[0] != tc.want {
			t.Errorf("Suggest(%q): got %v, want [%q]", tc.cmd, got, tc.want)
		}
	}
}

func TestGoTestNoDotDotDot(t *testing.T) {
	got := suggest("go test", 1)
	if len(got) == 0 || got[0] != "go test ./..." {
		t.Errorf("got %v", got)
	}
}

// ── make ─────────────────────────────────────────────────────────────────────

func TestMakeTargetTypo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte(
		"build:\n\tgo build ./...\n\ninstall:\n\tgo install ./...\n\ntest:\n\tgo test ./...\n",
	), 0644)

	c := ctx.ShellContext{WorkingDir: dir}
	got := suggestCtx("make insatll", 2, c)
	if len(got) == 0 || got[0] != "make install" {
		t.Errorf("got %v", got)
	}
}

func TestMakeBinaryTypo(t *testing.T) {
	got := suggest("mkae build", 127)
	if len(got) == 0 || got[0] != "make build" {
		t.Errorf("got %v", got)
	}
}

// ── typo (generic) ───────────────────────────────────────────────────────────

func TestGenericTypos(t *testing.T) {
	cases := []struct{ cmd, want string }{
		{"sl -la", "ls -la"},
		{"gti status", "git status"},
		{"dokcer ps", "docker ps"},
		{"claer", "clear"},
		{"gerp foo bar", "grep foo bar"},
		{"mkae all", "make all"},
	}
	for _, tc := range cases {
		got := suggest(tc.cmd, 127)
		if len(got) == 0 {
			t.Errorf("Suggest(%q): got no fix, want [%q]", tc.cmd, tc.want)
		} else if got[0] != tc.want {
			t.Errorf("Suggest(%q): got %q, want %q", tc.cmd, got[0], tc.want)
		}
	}
}

func TestSuggestEmpty(t *testing.T) {
	if got := suggest("", 1); len(got) != 0 {
		t.Errorf("expected nil for empty cmd, got %v", got)
	}
}
