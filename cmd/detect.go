/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
	"github.com/cage1016/alfred-video2ringtone/scripts"
)

// detectCmd represents the yr command
var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect URL from browser or clipboard",
	Run:   runDetectCmd,
}

func runDetectCmd(cmd *cobra.Command, args []string) {
	// check previous job
	p, _ := alfred.LoadOngoingProcess(wf)
	if p.Step != "" {
		wf.NewItem(p.Title).
			Subtitle(p.Process).
			Valid(false)

		count++
		wf.Rerun(1)
		wf.SendFeedback()
		return
	}

	// priority
	// 1. Video Page
	// 2. Clipboard
	clipboardFn := func() {
		query, _ := clipboard.ReadAll()
		if !lib.IsVideoURLValid(query) {
			wf.NewItem("No Valid Video URL found").
				Subtitle("Do not found a Valid Video URL in the clipboard").
				Largetype(alfred.GetSupportSites(wf)).
				Icon(DefaultDisabledIcon).
				Valid(false)
		} else {
			wf.NewItem(query).
				Subtitle("↩ To Convert Video 2 Ringtone").
				Valid(true).
				Arg(query)
		}
	}

	ccmd := exec.Command("osascript", scripts.Path("get_title_and_url.js"))
	out, err := ccmd.CombinedOutput()
	logrus.Debugf("out: %v", string(out))
	if err != nil {
		// 2. Clipboard
		// NOT browser as frontmost app
		clipboardFn()
	} else {
		// 1. Video Page
		// browser as frontmost app with url & title
		a := strings.Split(string(out), "\n")
		url, title := a[0], a[1]

		if !lib.IsVideoURLValid(url) {
			// 2. Clipboard
			clipboardFn()
		} else {
			wf.NewItem(title).
				Subtitle("↩ To Convert Video 2 Ringtone").
				Valid(true).
				Arg(strings.Join([]string{url, title}, "\n"))
		}
	}

	wf.SendFeedback()
}

func init() {
	rootCmd.AddCommand(detectCmd)
}