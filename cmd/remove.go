/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "A brief description of your command",
	Run:   runRemoveCmd,
}

func runRemoveCmd(cmd *cobra.Command, args []string) {
	data, _ := alfred.LoadOngoingRingTone(wf)

	if item, ok := data.Items[args[0]]; ok {
		fn := filepath.Join(alfred.GetOutput(wf), item.Name)
		if _, err := os.Stat(fn); err == nil {
			err = os.Remove(fn)
			if err == nil {
				// start notifier
				lib.Notifier(item.Name + "has been removed")
			}
		}
		delete(data.Items, args[0])
	}

	alfred.StoreOngoingRingTone(wf, data)
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
