package lib

import (
	"fmt"
	"os/exec"
	"strings"
)

const p = `tell application id "com.runningwithcrayons.Alfred" to run trigger "notifier" in workflow "com.kaichu.video2ringtone" with argument "%s"`
const title = "Video 2 Ringtone"

func Notifier(msg string) {
	exec.Command("osascript", "-e", fmt.Sprintf(p, strings.Join([]string{title, msg}, ","))).Run()
}
