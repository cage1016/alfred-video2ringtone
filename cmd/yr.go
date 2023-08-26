/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-yt2ringtone/alfred"
	"github.com/cage1016/alfred-yt2ringtone/scripts"
)

// yrCmd represents the yr command
var yrCmd = &cobra.Command{
	Use:   "yr",
	Short: "A brief description of your command",
	Run:   runYrCmd,
}

func runYrCmd(cmd *cobra.Command, args []string) {
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
	// 1. Universal Action
	// 2. Youtube Page
	// 3. Clipboard

	logrus.Debugf("args[0]: %v", args[0])
	if args[0] != "" {
		// 1. Universal Action
		url := args[0]
		if !IsYoutubeURLValid(url) {
			wf.NewItem("No Youtube URL found").
				Subtitle("Do not found a Youtube URL from Universal Action").
				Icon(DefaultDisabledIcon).
				Valid(false)
		} else {
			wf.NewItem(url).
				Subtitle("↩ To Convert Youtube 2 Ringtone").
				Valid(true).
				Arg(url).
				Var("action", "parse")

			// load exist ringtone
			ringtone, _ := alfred.LoadOngoingRingTone(wf)
			if rt, ok := ringtone.Item[url]; ok {
				p := filepath.Join(wf.DataDir(), rt.Name)
				t := time.Unix(rt.CreatedAt, 0).Local().Format("2006-01-02 15:04:05")
				wf.NewItem(rt.Name).
					Subtitle(fmt.Sprintf("^ ,↩ Action in Alfred %s. %s", t, rt.Info)).
					Valid(true).
					Quicklook(p).
					Largetype(fmt.Sprintf("Created At %s \n\n%s", t, rt.Info)).
					Icon(RingToneIcon).
					Arg(p).
					Var("action", "reveal").
					Ctrl().
					Subtitle("↩ Remove item").
					Valid(true).
					Var("action", "remove").
					Arg(url)
			}
		}
	} else {
		clipboardfn := func() {
			query, _ := clipboard.ReadAll()
			if !IsYoutubeURLValid(query) {
				wf.NewItem("No Youtube URL found").
					Subtitle("Do not found a Youtube URL in the clipboard").
					Icon(DefaultDisabledIcon).
					Valid(false)
			} else {
				wf.NewItem(query).
					Subtitle("↩ To Convert Youtube 2 Ringtone").
					Valid(true).
					Arg(query).
					Var("action", "parse")

				// load exist ringtone
				ringtone, _ := alfred.LoadOngoingRingTone(wf)
				if rt, ok := ringtone.Item[query]; ok {
					p := filepath.Join(wf.DataDir(), rt.Name)
					t := time.Unix(rt.CreatedAt, 0).Local().Format("2006-01-02 15:04:05")
					wf.NewItem(rt.Name).
						Subtitle(fmt.Sprintf("^ ,↩ Action in Alfred %s. %s", t, rt.Info)).
						Valid(true).
						Quicklook(p).
						Largetype(fmt.Sprintf("Created At %s \n\n%s", t, rt.Info)).
						Icon(RingToneIcon).
						Arg(p).
						Var("action", "reveal").
						Ctrl().
						Subtitle("↩ Remove item").
						Valid(true).
						Var("action", "remove").
						Arg(query)
				}
			}
		}

		ccmd := exec.Command("osascript", scripts.Path("get_title_and_url.js"))
		out, err := ccmd.CombinedOutput()
		logrus.Debugf("out: %v", string(out))
		if err != nil {
			// 3. Clipboard
			// NOT browser as frontmost app
			clipboardfn()
		} else {
			// 2. Youtube Page
			// browser as frontmost app with url & title
			a := strings.Split(string(out), "\n")
			url, title := a[0], a[1]

			if !IsYoutubeURLValid(url) {
				// 3. Clipboard
				clipboardfn()
			} else {
				wf.NewItem(title).
					Subtitle("↩ To Convert Youtube 2 Ringtone").
					Valid(true).
					Arg(strings.Join([]string{url, title}, "\n")).
					Var("action", "parse")

				// load exist ringtone
				ringtone, _ := alfred.LoadOngoingRingTone(wf)
				if rt, ok := ringtone.Item[url]; ok {
					p := filepath.Join(wf.DataDir(), rt.Name)
					t := time.Unix(rt.CreatedAt, 0).Local().Format("2006-01-02 15:04:05")
					wf.NewItem(rt.Name).
						Subtitle(fmt.Sprintf("^ ,↩ Action in Alfred %s. %s", t, rt.Info)).
						Valid(true).
						Quicklook(p).
						Largetype(fmt.Sprintf("Created At %s \n\n%s", t, rt.Info)).
						Icon(RingToneIcon).
						Arg(p).
						Var("action", "reveal").
						Ctrl().
						Subtitle("↩ Remove item").
						Valid(true).
						Var("action", "remove").
						Arg(url)
				}
			}
		}
	}

	wf.SendFeedback()
}

func init() {
	rootCmd.AddCommand(yrCmd)
}
