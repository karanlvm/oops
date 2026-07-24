package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

// knownCmds is the set of common shell commands used for Levenshtein fallback.
var knownCmds = []string{
	"ls", "ll", "la", "cd", "pwd", "mkdir", "rmdir", "rm", "cp", "mv",
	"cat", "less", "more", "head", "tail", "grep", "awk", "sed", "cut",
	"sort", "uniq", "wc", "tr", "tee", "find", "xargs", "which", "type",
	"echo", "printf", "read", "export", "source", "alias", "unalias",
	"history", "clear", "reset", "exit", "logout",
	"touch", "chmod", "chown", "chgrp", "ln", "stat",
	"ps", "top", "htop", "kill", "killall", "pkill", "pgrep", "jobs", "bg", "fg",
	"df", "du", "free", "uname", "whoami", "id", "who", "w",
	"date", "cal", "uptime",
	"ssh", "scp", "rsync", "curl", "wget", "ping", "traceroute", "netstat", "ss",
	"tar", "gzip", "gunzip", "zip", "unzip", "bzip2", "xz",
	"sudo", "su", "env", "printenv", "set",
	"man", "info", "help",
	"git", "make", "cmake",
	"python3", "python", "pip3", "pip",
	"node", "npm", "npx", "yarn", "pnpm",
	"go", "java", "javac", "mvn", "gradle",
	"ruby", "gem", "bundle",
	"cargo", "rustc", "rustup",
	"docker", "docker-compose", "kubectl", "helm",
	"brew", "apt", "apt-get", "yum", "dnf", "pacman", "snap",
	"systemctl", "service", "journalctl",
	"vim", "vi", "nano", "emacs",
	"tmux", "screen",
	"jq", "yq", "fzf", "ripgrep", "rg", "fd", "bat", "eza", "exa",
	"terraform", "ansible", "vault",
	"openssl", "gpg", "ssh-keygen",
}

// directSubs are exact-match substitutions for the most common typos.
// These fire before Levenshtein so they can be intentionally different
// (e.g. bare "python" → "python3" even though distance is 1 char suffix).
var directSubs = map[string]string{
	// ls
	"sl": "ls", "ks": "ls", "lss": "ls", "lls": "ls", "lsa": "ls -a",
	// cd
	"dc": "cd", "ccd": "cd",
	// python — handle the python→python3 alias ambiguity explicitly
	"python":  "python3",
	"pyhton":  "python3",
	"pytohn":  "python3",
	"pyton":   "python3",
	"pthon":   "python3",
	"pythno":  "python3",
	"pythoon": "python3",
	"pytho":   "python3",
	"phyton":  "python3",
	// pip
	"pip":  "pip3",
	"piip": "pip3",
	"pipp": "pip3",
	// node/npm
	"noed": "node",
	"ndoe": "node",
	"nom":  "npm",
	"nmp":  "npm",
	// clear
	"claer": "clear",
	"clera": "clear",
	"lcear": "clear",
	// grep
	"gerp": "grep",
	"gep":  "grep",
	"grpe": "grep",
	// cat
	"cta": "cat",
	"act": "cat",
	// mkdir
	"mkdr":  "mkdir",
	"mdkir": "mkdir",
	"mkidr": "mkdir",
	// rm
	"mr": "rm",
	// mv
	"vm": "mv",
	// cp
	"pc": "cp",
	// touch
	"toch":  "touch",
	"tuoch": "touch",
	"touhc": "touch",
	// sudo
	"suod": "sudo",
	"sduo": "sudo",
	"udo":  "sudo",
	"sudi": "sudo",
	// man
	"amn": "man",
	"mna": "man",
	// curl
	"crul":  "curl",
	"clur":  "curl",
	"cukl":  "curl",
	"ckurl": "curl",
	// ssh
	"shh": "ssh",
	"ssh ": "ssh",
	// docker
	"dokcer": "docker",
	"dcoker": "docker",
	"docekr": "docker",
	"doker":  "docker",
	// kubectl
	"kubctl":   "kubectl",
	"kubeclt":  "kubectl",
	"kubectll": "kubectl",
	"kube":     "kubectl",
	// vim
	"vmi": "vim",
	"ivm": "vim",
	// nano
	"naon": "nano",
	// less
	"lses": "less",
	"lesd": "less",
	// more
	"mroe": "more",
	// history
	"hisotry": "history",
	"histroy": "history",
	"hsitory": "history",
	// echo
	"ehco": "echo",
	"ecoh": "echo",
	// export
	"exprot": "export",
	"exoprt": "export",
	// source
	"soruce":  "source",
	"sourc":   "source",
	"srouce":  "source",
	"souurce": "source",
	// which
	"whihc": "which",
	"whcih": "which",
	"wihch": "which",
	// pwd
	"wpd": "pwd",
	"dpw": "pwd",
	"pdw": "pwd",
	// chmod
	"chmo":  "chmod",
	"chmdo": "chmod",
	// chown
	"choown": "chown",
	"chnow":  "chown",
	// find
	"fidn": "find",
	"fnid": "find",
	"fnd":  "find",
	// xargs
	"xarg":  "xargs",
	"xarsg": "xargs",
	// awk
	"akw": "awk",
	"wak": "awk",
	// sed
	"esd": "sed",
	"dse": "sed",
	// tar
	"tra": "tar",
	"art": "tar",
	// unzip
	"unizp": "unzip",
	"unzpi": "unzip",
	// ps
	"sp": "ps",
	// kill
	"klil":  "kill",
	"kil":   "kill",
	"klill": "kill",
	// top
	"tpo": "top",
	"opt": "top",
	// df
	"fd": "df",
	// du
	"ud": "du",
	// wc
	"cw": "wc",
	// head
	"haed": "head",
	"eahd": "head",
	// tail
	"tial":  "tail",
	"tlai":  "tail",
	"tlail": "tail",
	// sort
	"srot":  "sort",
	"srto":  "sort",
	"osrt":  "sort",
	// uniq
	"uinq": "uniq",
	"inuq": "uniq",
	// cut
	"uct": "cut",
	// make
	"mkae": "make",
	"amke": "make",
	"maek": "make",
	"mak":  "make",
	// cmake
	"camke":  "cmake",
	"cmkae":  "cmake",
	"cmakee": "cmake",
	// go
	"og": "go",
	// java
	"jvaa": "java",
	"avaj": "java",
	"ajva": "java",
	// javac
	"javaac": "javac",
	"jaavac": "javac",
	// mvn
	"mnv": "mvn",
	"vmn": "mvn",
	"nvm": "mvn",
	// gradle
	"gradel":  "gradle",
	"gladle":  "gradle",
	"gradele": "gradle",
	// cargo
	"cargoo": "cargo",
	"craog":  "cargo",
	"carog":  "cargo",
	// rustc
	"rustcc": "rustc",
	"rsutc":  "rustc",
	// ruby
	"rubye": "ruby",
	"rubby": "ruby",
	"rbuy":  "ruby",
	// gem
	"gme": "gem",
	"emg": "gem",
	// bundle
	"undle":  "bundle",
	"bundel": "bundle",
	"bunlde": "bundle",
	// brew
	"bre":  "brew",
	"bew":  "brew",
	"brwe": "brew",
	"breu": "brew",
	// apt
	"apt-gt": "apt-get",
	"at-get": "apt-get",
	"apt-ge": "apt-get",
	// yum
	"ymu": "yum",
	"uym": "yum",
	// dnf
	"dfn": "dnf",
	"ndf": "dnf",
	// pacman
	"pacamn": "pacman",
	"pamcan": "pacman",
	// snap
	"snpa": "snap",
	"sanp": "snap",
	// systemctl
	"systmectl":  "systemctl",
	"systemclt":  "systemctl",
	"systemcttl": "systemctl",
	"sysctl":     "systemctl",
	// service
	"sevice":  "service",
	"serivce": "service",
	// journalctl
	"journalclt":  "journalctl",
	"jounalctl":   "journalctl",
	"journalcttl": "journalctl",
	// terraform
	"terraofrm": "terraform",
	"terrafrom": "terraform",
	"terrafrorm": "terraform",
	// ansible
	"ansibel":  "ansible",
	"ansiblle": "ansible",
	// helm
	"hlme": "helm",
	"heml": "helm",
	// rsync
	"rysnc": "rsync",
	"rscyn": "rsync",
	// wget
	"wegt": "wget",
	"wtge": "wget",
	// ping
	"pign": "ping",
	"pnig": "ping",
	// ssh-keygen
	"ssh-keygne": "ssh-keygen",
	// tmux
	"tmxu": "tmux",
	"tmuux": "tmux",
	// screen
	"sceren": "screen",
	"screne": "screen",
	// stat
	"stta": "stat",
	"tsat": "stat",
	// ln
	"nl": "ln",
	// chgrp
	"chgrps": "chgrp",
	// uname
	"unaem": "uname",
	"nmae":  "uname",
	// whoami
	"whoamii": "whoami",
	"whomia":  "whoami",
	"whaomi":  "whoami",
	// date
	"dtae": "date",
	"daet": "date",
	// uptime
	"uptiem": "uptime",
	"uptme":  "uptime",
	// exit
	"exti": "exit",
	"etix": "exit",
	// git (so bare "gti" works)
	"gti":  "git",
	"gigt": "git",
	"gitt": "git",
	"geet": "git",
	"goit": "git",
	"guit": "git",
}

type typoRule struct{}

func (r *typoRule) Name() string { return "typo" }

func (r *typoRule) Match(cmd string, exitCode int, ctx context.ShellContext) bool {
	// Only fire on "command not found" (127) or generic failure (1, 2)
	return exitCode == 127 || exitCode == 1 || exitCode == 2
}

func (r *typoRule) Fix(cmd string, exitCode int, ctx context.ShellContext) []string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = " " + strings.Join(parts[1:], " ")
	}

	// 1. Direct substitution table
	if sub, ok := directSubs[name]; ok {
		return []string{sub + rest}
	}

	// 2. Levenshtein fallback — only suggest if close enough
	if best, dist := closestMatch(name, knownCmds); dist == 1 {
		return []string{best + rest}
	}

	return nil
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	row := make([]int, len(b)+1)
	for j := range row {
		row[j] = j
	}
	for i, ra := range a {
		prev := i + 1
		for j, rb := range b {
			next := row[j]
			if ra != rb {
				if row[j+1]+1 < next+1 {
					next = row[j+1] + 1
				} else {
					next = next + 1
				}
				if prev+1 < next {
					next = prev + 1
				}
			}
			row[j] = prev
			prev = next
		}
		row[len(b)] = prev
	}
	return row[len(b)]
}

// closestMatch returns the closest candidate and its edit distance.
func closestMatch(s string, candidates []string) (string, int) {
	best, bestDist := "", len(s)+1
	for _, c := range candidates {
		d := levenshtein(s, c)
		if d < bestDist {
			bestDist, best = d, c
		}
	}
	return best, bestDist
}
