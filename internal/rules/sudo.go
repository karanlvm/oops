package rules

import (
	"strings"

	"github.com/karanlvm/oops/internal/context"
)

// Commands that almost always require root on Linux/macOS.
var needsRoot = map[string]bool{
	// package managers
	"apt": true, "apt-get": true, "apt-cache": true, "apt-key": true,
	"yum": true, "dnf": true, "rpm": true, "zypper": true,
	"pacman": true, "pamac": true,
	"dpkg": true, "dpkg-reconfigure": true,
	"snap": true,
	"update-alternatives": true, "update-rc.d": true,
	// init / services
	"systemctl": true, "service": true, "chkconfig": true,
	"journalctl": true, "loginctl": true,
	// disk / fs
	"mount": true, "umount": true, "fdisk": true, "gdisk": true,
	"parted": true, "mkfs": true, "fsck": true, "e2fsck": true,
	"blkid": true, "lsblk": true, "cryptsetup": true,
	"mdadm": true, "lvdisplay": true, "pvdisplay": true, "vgdisplay": true,
	"lvcreate": true, "vgcreate": true, "pvcreate": true,
	// networking
	"iptables": true, "ip6tables": true, "nftables": true,
	"ufw": true, "firewall-cmd": true,
	"ip": true, "ifconfig": true, "iwconfig": true,
	"tc": true, "brctl": true,
	// users / auth
	"useradd": true, "usermod": true, "userdel": true,
	"groupadd": true, "groupmod": true, "groupdel": true,
	"passwd": true, "chpasswd": true, "visudo": true,
	"adduser": true, "deluser": true, "addgroup": true, "delgroup": true,
	// kernel / hardware
	"modprobe": true, "rmmod": true, "insmod": true,
	"sysctl": true, "dmesg": true,
	"lshw": true, "dmidecode": true,
	// power
	"reboot": true, "shutdown": true, "halt": true, "poweroff": true, "init": true,
	// nginx / web servers (common dev scenario)
	"nginx": true, "apache2": true, "httpd": true,
	// other
	"tcpdump": true, "wireshark": true, "strace": true,
	"crontab": true,
}

type sudoRule struct{}

func (r *sudoRule) Name() string { return "sudo" }

func (r *sudoRule) Match(cmd string, exitCode int, _ context.ShellContext) bool {
	if strings.HasPrefix(cmd, "sudo ") {
		return false
	}
	name := first(cmd)
	// 126 = found but not executable (permission denied)
	if exitCode == 126 {
		return true
	}
	// 1 = generic failure for many commands that need root
	if exitCode == 1 && needsRoot[name] {
		return true
	}
	return false
}

func (r *sudoRule) Fix(cmd string, exitCode int, _ context.ShellContext) []string {
	// ./script.sh with permission denied → chmod +x it first
	parts := strings.Fields(cmd)
	if len(parts) > 0 && strings.HasPrefix(parts[0], "./") && exitCode == 126 {
		return []string{
			"chmod +x " + parts[0],
			cmd,
		}
	}
	return []string{"sudo " + cmd}
}
