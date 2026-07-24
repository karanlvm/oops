package rules

import (
	"fmt"
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

// All known git subcommands — used for Levenshtein fallback.
var knownGitCmds = []string{
	"add", "am", "annotate", "apply", "archive", "bisect", "blame",
	"branch", "bundle", "checkout", "cherry", "cherry-pick", "clean",
	"clone", "commit", "config", "describe", "diff", "fetch",
	"format-patch", "gc", "grep", "gui", "init", "log", "ls-files",
	"ls-remote", "ls-tree", "merge", "merge-base", "mergetool", "mv",
	"notes", "pull", "push", "rebase", "reflog", "remote", "reset",
	"restore", "revert", "rm", "shortlog", "show", "show-branch",
	"sparse-checkout", "stash", "status", "submodule", "switch",
	"tag", "worktree",
}

// gitVerbTypos maps common git subcommand typos to their corrections.
var gitVerbTypos = map[string]string{
	// push
	"psuh":  "push",
	"pish":  "push",
	"pusj":  "push",
	"puch":  "push",
	"puhs":  "push",
	"pus":   "push",
	"ush":   "push",
	"puhsh": "push",
	"puush": "push",
	"ppush": "push",
	"pushh": "push",
	"pussh": "push",
	"pugh":  "push",
	"puse":  "push",
	"puis":  "push",
	// pull
	"pul":  "pull",
	"plul": "pull",
	"pll":  "pull",
	"upll": "pull",
	"pukl": "pull",
	// commit
	"comit":    "commit",
	"comitt":   "commit",
	"commot":   "commit",
	"commmit":  "commit",
	"coommit":  "commit",
	"commiy":   "commit",
	"ocmmit":   "commit",
	"commiit":  "commit",
	"commi":    "commit",
	"comimt":   "commit",
	"committ":  "commit",
	"commmitt": "commit",
	"commti":   "commit",
	"cmmoit":   "commit",
	"cmomit":   "commit",
	"ommit":    "commit",
	// checkout
	"chekcout":  "checkout",
	"checokut":  "checkout",
	"chekout":   "checkout",
	"checout":   "checkout",
	"checkotu":  "checkout",
	"cehckout":  "checkout",
	"cechkout":  "checkout",
	"chwckout":  "checkout",
	"chcekout":  "checkout",
	"checkou":   "checkout",
	"ckeckout":  "checkout",
	"chekckout": "checkout",
	"checkot":   "checkout",
	"checkiut":  "checkout",
	"chekotu":   "checkout",
	"chekcotu":  "checkout",
	"chceckout": "checkout",
	// status
	"statsu":  "status",
	"sttaus":  "status",
	"staus":   "status",
	"satuts":  "status",
	"stauts":  "status",
	"satus":   "status",
	"stautus": "status",
	"statues": "status",
	"statuss": "status",
	"ststu":   "status",
	"sttus":   "status",
	"atatus":  "status",
	"tsatus":  "status",
	// branch
	"brnach":  "branch",
	"branh":   "branch",
	"branhc":  "branch",
	"brach":   "branch",
	"barnch":  "branch",
	"bracnh":  "branch",
	"bnarch":  "branch",
	"braanch": "branch",
	"brancch": "branch",
	"banrch":  "branch",
	"brnahc":  "branch",
	"brnch":   "branch",
	"rbench":  "branch",
	// merge
	"meger":  "merge",
	"mrege":  "merge",
	"mereg":  "merge",
	"megre":  "merge",
	"emrge":  "merge",
	"merg":   "merge",
	"meerge": "merge",
	"merrge": "merge",
	"mrge":   "merge",
	// fetch
	"feth":   "fetch",
	"ftch":   "fetch",
	"etch":   "fetch",
	"fecth":  "fetch",
	"fetxh":  "fetch",
	"ftetch": "fetch",
	"fetcch": "fetch",
	"fetc":   "fetch",
	"fetsh":  "fetch",
	// diff
	"dfif":  "diff",
	"dif":   "diff",
	"diif":  "diff",
	"dff":   "diff",
	"ddiff": "diff",
	"difff": "diff",
	"dffif": "diff",
	"fdif":  "diff",
	"dfi":   "diff",
	// log
	"lgo":   "log",
	"loog":  "log",
	"olg":   "log",
	"loggg": "log",
	"llgo":  "log",
	// rebase
	"reabse":  "rebase",
	"rabase":  "rebase",
	"rebsae":  "rebase",
	"rebae":   "rebase",
	"rbase":   "rebase",
	"rebas":   "rebase",
	"ebase":   "rebase",
	"rebaise": "rebase",
	// stash
	"stsh":   "stash",
	"stsah":  "stash",
	"satsh":  "stash",
	"sthas":  "stash",
	"tsash":  "stash",
	"saths":  "stash",
	"sttash": "stash",
	"stassh": "stash",
	"astsh":  "stash",
	"stahs":  "stash",
	// clone
	"cloen":  "clone",
	"colen":  "clone",
	"clne":   "clone",
	"clonee": "clone",
	"clonr":  "clone",
	"cloe":   "clone",
	"colon":  "clone",
	"lcone":  "clone",
	// add
	"addd": "add",
	"aad":  "add",
	"dda":  "add",
	"aedd": "add",
	"aadd": "add",
	// switch
	"swich":  "switch",
	"swtich": "switch",
	"swithc": "switch",
	"siwtch": "switch",
	"swticj": "switch",
	"siwthc": "switch",
	// restore
	"restroe": "restore",
	"resote":  "restore",
	"restor":  "restore",
	"resotre": "restore",
	"resore":  "restore",
	// init
	"inti":  "init",
	"iint":  "init",
	"nit":   "init",
	"initt": "init",
	"initi": "init",
	// show
	"shwo": "show",
	"sohw": "show",
	"hsow": "show",
	"shw":  "show",
	"shpw": "show",
	// reset
	"resot":  "reset",
	"rset":   "reset",
	"rste":   "reset",
	"rseet":  "reset",
	"reest":  "reset",
	"resett": "reset",
	"erset":  "reset",
	// tag
	"tga":  "tag",
	"atg":  "tag",
	"tagg": "tag",
	"targ": "tag",
	// cherry-pick
	"cherry-pik":  "cherry-pick",
	"cherrry-pick": "cherry-pick",
	"cherry-pic":  "cherry-pick",
	"cherrypick":  "cherry-pick",
	"cherry-pck":  "cherry-pick",
	"cheery-pick": "cherry-pick",
	// blame
	"blme":  "blame",
	"blaem": "blame",
	"blam":  "blame",
	"blsme": "blame",
	// remote
	"reomte":  "remote",
	"remotee": "remote",
	"remtoe":  "remote",
	"rmote":   "remote",
	"rmeote":  "remote",
	// config
	"configg": "config",
	"cofnig":  "config",
	"conifg":  "config",
	"cnofig":  "config",
	"configu": "config",
	// grep
	"grpe": "grep",
	"grap": "grep",
	"gerp": "grep",
	"gep":  "grep",
	// apply
	"appyl": "apply",
	"appl":  "apply",
	"aaply": "apply",
	// rm
	"mr": "rm",
	// mv (git mv)
	"vm": "mv",
	// shortlog
	"shrotlog": "shortlog",
	"shortlgo": "shortlog",
	// submodule
	"submoule":   "submodule",
	"submdoule":  "submodule",
	"submodul":   "submodule",
	"submoduele": "submodule",
	// worktree
	"worktre":   "worktree",
	"worktee":   "worktree",
	"worketree": "worktree",
	// reflog
	"reflgo":  "reflog",
	"reeflog": "reflog",
	// format-patch
	"formt-patch":  "format-patch",
	"format-ptach": "format-patch",
	// ls-files
	"ls-fles":   "ls-files",
	"ls-filles": "ls-files",
	// sparse-checkout
	"sparse-chekout": "sparse-checkout",
	// describe
	"descirbe": "describe",
	"descrbie": "describe",
	// annotate
	"anntoate": "annotate",
	"anntate":  "annotate",
	// bisect
	"bisecte": "bisect",
	"bisectt": "bisect",
	// bundle
	"bndule": "bundle",
	"bundel": "bundle",
	// archive
	"archiev": "archive",
	"archvie": "archive",
	// notes
	"ntoes": "notes",
	"noets": "notes",
	// gc
	"cg": "gc",
	// pull (alias variants)
	"puull": "pull",
	"ppull": "pull",
}

type gitRule struct{}

func (r *gitRule) Name() string { return "git" }

func (r *gitRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	f := first(cmd)
	return f == "git" || f == "gti" || f == "gitt" || f == "geet" || f == "goit"
}

func (r *gitRule) Fix(cmd string, exitCode int, ctx context.ShellContext) []string {
	parts := strings.Fields(cmd)

	// bare git binary typo with no subcommand
	if len(parts) == 1 {
		if parts[0] != "git" {
			return []string{"git"}
		}
		return nil
	}

	// Fix the binary name if it was a typo
	if parts[0] != "git" {
		parts[0] = "git"
	}

	verb := parts[1]
	rest := ""
	if len(parts) > 2 {
		rest = " " + strings.Join(parts[2:], " ")
	}

	// ── 1. Typo lookup table ─────────────────────────────────────────────────
	if correct, ok := gitVerbTypos[verb]; ok {
		return []string{fmt.Sprintf("git %s%s", correct, rest)}
	}

	// ── 2. Levenshtein fallback ──────────────────────────────────────────────
	if best, dist := closestMatch(verb, knownGitCmds); dist <= 2 && dist > 0 {
		return []string{fmt.Sprintf("git %s%s", best, rest)}
	}

	// ── 3. Semantic rules (verb is correctly spelled) ─────────────────────────

	// git push with no upstream → add -u origin <branch>
	if verb == "push" && exitCode == 128 && len(parts) == 2 {
		branch := ctx.GitBranch
		if branch == "" {
			branch = "main"
		}
		return []string{fmt.Sprintf("git push -u origin %s", branch)}
	}

	// git push rejected (non-fast-forward)
	if verb == "push" && (exitCode == 1 || exitCode == 128) && len(parts) >= 4 {
		return []string{
			fmt.Sprintf("git pull --rebase %s %s", parts[2], parts[3]),
			strings.Join(parts, " "),
		}
	}

	// git pull with no upstream
	if verb == "pull" && exitCode == 128 && len(parts) == 2 {
		branch := ctx.GitBranch
		if branch == "" {
			branch = "main"
		}
		return []string{fmt.Sprintf("git pull origin %s", branch)}
	}

	// git commit "message" (forgot -m)
	if verb == "commit" && len(parts) >= 3 && !strings.HasPrefix(parts[2], "-") {
		return []string{fmt.Sprintf("git commit -m %s", parts[2])}
	}

	// git branch -d unmerged → -D
	if verb == "branch" && exitCode != 0 {
		for i, p := range parts {
			if p == "-d" {
				fixed := make([]string, len(parts))
				copy(fixed, parts)
				fixed[i] = "-D"
				return []string{strings.Join(fixed, " ")}
			}
		}
	}

	// git checkout -b / git switch -c on existing branch → drop create flag
	if (verb == "checkout" || verb == "switch") && exitCode != 0 {
		for i, p := range parts {
			if p == "-b" || p == "-B" || p == "-c" || p == "-C" {
				fixed := append(append([]string{}, parts[:i]...), parts[i+1:]...)
				return []string{strings.Join(fixed, " ")}
			}
		}
	}

	// git add -p on untracked → plain git add
	if verb == "add" && containsFlag(parts, "-p", "--patch") && exitCode != 0 {
		var filtered []string
		for _, p := range parts[2:] {
			if p != "-p" && p != "--patch" {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) > 0 {
			return []string{"git add " + strings.Join(filtered, " ")}
		}
	}

	// git stash pop when nothing to pop
	if verb == "stash" && len(parts) >= 3 && parts[2] == "pop" && exitCode != 0 {
		return []string{"git stash list"}
	}

	// git clean without -f
	if verb == "clean" && exitCode != 0 && !containsFlag(parts, "-f", "--force") {
		return []string{strings.Join(parts, " ") + " -f"}
	}

	// git rebase conflict → offer abort
	if verb == "rebase" && exitCode != 0 && len(parts) == 2 {
		return []string{"git rebase --abort"}
	}

	// git merge conflict → offer abort
	if verb == "merge" && exitCode != 0 && len(parts) == 2 {
		return []string{"git merge --abort"}
	}

	// git log with bad flags → safe fallback
	if verb == "log" && exitCode != 0 && len(parts) > 2 {
		return []string{"git log --oneline -20"}
	}

	return nil
}

func containsFlag(parts []string, flags ...string) bool {
	for _, p := range parts {
		for _, f := range flags {
			if p == f {
				return true
			}
		}
	}
	return false
}
