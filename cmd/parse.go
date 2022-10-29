/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>

*/
package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-yt2ringtone/alfred"
)

var rangeRegex = regexp.MustCompile(`(?m)^[0-2][0-3]:[0-5][0-9]:[0-5][0-9](?:,(40|[0-3][0-9]|[12][0-9]|[1-9]))?(?:,(40|[0-3][0-9]|[12][0-9]|[1-9])){0,2}$`)
var ytURLRegex = regexp.MustCompile(`(?m)^(https?\:\/\/)?(www\.youtube\.com|youtu\.be)\/.+$`)

func IsRangeValid(s string) bool {
	return rangeRegex.MatchString(s)
}

func IsYoutubeURLValid(s string) bool {
	return ytURLRegex.MatchString(s)
}

// parseCmd represents the parse command
var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse ringtone duration, fadeIn, fadeOut",
	Run:   runParseCmd,
}

var count = 0

func runParseCmd(cmd *cobra.Command, args []string) {
	CheckForUpdate()

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

	input, _ := cmd.Flags().GetString("input")
	if input == "" {
		wf.NewItem("No input").
			Subtitle("Please provide youtube url").
			Icon(DefaultDisabledIcon).
			Valid(true)
		wf.SendFeedback()
		return
	}

	var url, title string
	buf := strings.Split(input, "\n")
	if len(buf) == 2 {
		url, title = buf[0], buf[1]
	} else if len(buf) == 1 {
		url = buf[0]
		title = buf[0]
	}

	if !IsYoutubeURLValid(url) {
		wf.NewItem(fmt.Sprintf("\"%s\" is invalid Youtube URL", url)).
			Subtitle("Try another query?").
			Icon(DefaultDisabledIcon).
			Valid(false)
		wf.SendFeedback()
		return
	}

	help := alfred.GetHelp(wf)
	if len(args) == 0 || !IsRangeValid(args[0]) {
		wf.NewItem(title).
			Subtitle("⌘+L for help. e.g. \"HH:MM:SS | HH:MM:SS,duration | HH:MM:SS,duration,FadeIn&Out | HH:MM:SS,duration,FadeIn,FadeOut\"").
			Quicklook(url).
			Largetype(help).
			Icon(DefaultDisabledIcon).
			Valid(false)
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
			wf.NewItem(title).
				Subtitle(fmt.Sprintf("⌘+L for help. Then duration %ss must >= (fadeIn %ss + fadeOut %ss)", t, fin, fout)).
				Arg(fmt.Sprintf("%s,%s,%s,%s", ss, t, fin, fout)).
				Quicklook(url).
				Largetype(help).
				Icon(DefaultDisabledIcon).
				Valid(false)
		} else {
			wf.NewItem(title).
				Subtitle(fmt.Sprintf("⌘+L for help. Start %s with %ss duration, %ss fadeIn, %ss fadeOut", ss, t, fin, fout)).
				Arg(fmt.Sprintf("%s,%s,%s,%s", ss, t, fin, fout)).
				Quicklook(url).
				Largetype(help).
				Valid(true)
		}
	}

	wf.SendFeedback()
}

func init() {
	rootCmd.AddCommand(parseCmd)
	parseCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
