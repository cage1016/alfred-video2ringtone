package lib

import (
	"fmt"
	"os/exec"
	"strings"
)

const title = "Video 2 Ringtone"

const notifier = `tell application id "com.runningwithcrayons.Alfred" to run trigger "notifier" in workflow "com.kaichu.video2ringtone" with argument "%s"`

func Notifier(msg string) {
	exec.Command("osascript", "-e", fmt.Sprintf(notifier, strings.Join([]string{title, msg}, ","))).Run()
}

const detect = `tell application id "com.runningwithcrayons.Alfred" to run trigger "detect" in workflow "com.kaichu.video2ringtone" with argument "%s"`

func Detect(input string) {
	exec.Command("osascript", "-e", fmt.Sprintf(detect, input)).Run()
}
