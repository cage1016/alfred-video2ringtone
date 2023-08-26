/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "A brief description of your command",
	Run:   runRemoveCmd,
}

func runRemoveCmd(cmd *cobra.Command, args []string) {
	data, _ := alfred.LoadOngoingRingTone(wf)

	if item, ok := data.Item[args[0]]; ok {
		fn := filepath.Join(wf.DataDir(), item.Name)
		if _, err := os.Stat(fn); err == nil {
			err = os.Remove(fn)
			if err == nil {
				// start notifier
				notifier("YouTube 2 Ringtone", item.Name+"has been removed")
			}
		}
		delete(data.Item, args[0])
	}

	alfred.StoreOngoingRingTone(wf, data)
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
