/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
)

// ylCmd represents the yl command
var ylCmd = &cobra.Command{
	Use:   "yl",
	Short: "A brief description of your command",
	Run:   runYlCmd,
}

func runYlCmd(c *cobra.Command, args []string) {
	data, _ := alfred.LoadOngoingRingTone(wf)
	for url, rt := range data.Item {
		p := filepath.Join(wf.DataDir(), rt.Name)
		t := time.Unix(rt.CreatedAt, 0).Local().Format("2006-01-02 15:04:05")
		uid := strconv.FormatInt(rt.CreatedAt, 10)

		ni := wf.NewItem(rt.Name).
			Subtitle(fmt.Sprintf("⌥ ^ ,↩ Action in Alfred %s. %s", t, rt.Info)).
			Valid(true).
			Quicklook(p).
			Largetype(fmt.Sprintf("Created At %s \n\n%s", t, rt.Info)).
			Icon(RingToneIcon).
			UID(uid).
			Arg(p)

		ni.Opt().
			Subtitle("↩ Re convert again").
			Valid(true).
			Arg(url)

		ni.Ctrl().
			Subtitle("↩ Remove Item").
			Valid(true).
			Arg(url)
	}

	if args[0] != "" {
		wf.Filter(args[0])
	}
	wf.WarnEmpty("No matching items", "Try a different query?")
	wf.SendFeedback()
}

func init() {
	rootCmd.AddCommand(ylCmd)
}
