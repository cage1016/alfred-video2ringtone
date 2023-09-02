/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
)

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse ringtone duration, fadeIn, fadeOut",
	Run:   runParseCmd,
}

var count = 0

func runParseCmd(cmd *cobra.Command, args []string) {
	// check if previous job is running and show it
	if p, _ := alfred.LoadOngoingProcess(wf); p.Step != "" {
		wf.NewItem(p.Title).
			Subtitle(p.Process).
			Valid(false)

		count++
		wf.Rerun(1)
		wf.SendFeedback()
		return
	}

	// parse url and title from input
	var url, title string
	if input, _ := cmd.Flags().GetString("input"); input == "" {
		wf.NewItem("No input").
			Subtitle("Please provide Video url").
			Icon(VideoLinkDisabledIcon).
			Valid(true)
		wf.SendFeedback()
		return
	} else {
		buf := strings.Split(input, "\n")
		if len(buf) == 2 {
			url, title = buf[0], buf[1]
		} else if len(buf) == 1 {
			url = buf[0]
			title = buf[0]
		}
	}

	// check if url is valid
	if !lib.IsVideoURLValid(url) {
		wf.NewItem(fmt.Sprintf("\"%s\" is invalid Video URL", url)).
			Subtitle("Try another query?").
			Icon(VideoLinkDisabledIcon).
			Valid(false)
		wf.SendFeedback()
		return
	}

	help := alfred.GetHelp(wf)
	if len(args) == 0 || !lib.IsRangeValid(args[0]) {
		wi := wf.NewItem(title).
			Subtitle("⌘+L ⌥, e.g. \"HH:MM:SS | HH:MM:SS,duration | HH:MM:SS,duration,FadeIn&Out | HH:MM:SS,duration,FadeIn,FadeOut\"").
			Quicklook(url).
			Largetype(help).
			Icon(VideoLinkDisabledIcon).
			Valid(false)

		wi.Opt().
			Subtitle(fmt.Sprintf("↩ Open %s", url)).
			Arg(url).
			Valid(true).
			Var("action", "open-url")
	} else {
		buf := strings.Split(args[0], ",")
		dt, dfin, dfout := "40", "3", "3"
		var ss, t, fin, fout string
		switch l := len(buf); l {
		case 1:
			// 00:00:00
			ss, t, fin, fout = buf[0], dt, dfin, dfout
		case 2:
			// 00.00.00,40
			ss, t, fin, fout = buf[0], buf[1], dfin, dfout
		case 3:
			// 00.00.00,40,3
			ss, t, fin, fout = buf[0], buf[1], buf[2], buf[2]
		case 4:
			// 00.00.00,40,3,3
			ss, t, fin, fout = buf[0], buf[1], buf[2], buf[3]
		}

		it, _ := strconv.Atoi(t)
		ifin, _ := strconv.Atoi(fin)
		ifout, _ := strconv.Atoi(fout)

		if it < (ifin + ifout) {
			wi := wf.NewItem(title).
				Subtitle(fmt.Sprintf("⌘+L ⌥, Then duration %ss must >= (fadeIn %ss + fadeOut %ss)", t, fin, fout)).
				Arg(fmt.Sprintf("%s,%s,%s,%s", ss, t, fin, fout)).
				Quicklook(url).
				Largetype(help).
				Icon(VideoLinkDisabledIcon).
				Valid(false)

			wi.Opt().
				Subtitle(fmt.Sprintf("↩ Open %s", url)).
				Arg(url).
				Valid(true).
				Var("action", "open-url")
		} else {
			wi := wf.NewItem(title).
				Subtitle(fmt.Sprintf("⌘+L ⌥, Start %s with %ss duration, %ss fadeIn, %ss fadeOut", ss, t, fin, fout)).
				Arg(fmt.Sprintf("%s,%s,%s,%s", ss, t, fin, fout)).
				Quicklook(url).
				Largetype(help).
				Icon(VideoLinkIcon).
				Valid(true).
				Var("action", "convert")

			wi.Opt().
				Subtitle(fmt.Sprintf("↩ Open %s", url)).
				Arg(url).
				Valid(true).
				Var("action", "open-url")
		}
	}

	// load exist ringtone
	ringtone, _ := alfred.LoadOngoingRingTone(wf)
	if rt, ok := ringtone.Items[url]; ok {
		p := filepath.Join(alfred.GetOutput(wf), rt.Name)
		t := time.Unix(rt.CreatedAt, 0).Local().Format("2006-01-02 15:04:05")
		wi := wf.NewItem(rt.Name).
			Subtitle(fmt.Sprintf("⌥ ,↩ Action in Alfred %s. %s", t, rt.Info)).
			Valid(true).
			Quicklook(p).
			Largetype(fmt.Sprintf("Created At %s \n\n%s", t, rt.Info)).
			Icon(RingToneIcon).
			Arg(p).
			Var("action", "alfred-action")

		wi.Opt().
			Subtitle("↩ Remove Item").
			Valid(true).
			Arg(url).
			Var("action", "remove")
	}

	wf.SendFeedback()
}

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
