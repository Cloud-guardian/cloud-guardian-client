package linux_reboot

import (
	"os/exec"
)

func Reboot() error {
	return exec.Command("reboot").Run()
}
