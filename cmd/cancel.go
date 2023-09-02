/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
)

// cancelCmd represents the cancel command
var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "A brief description of your command",
	Run:   runCancelCmd,
}

func runCancelCmd(cmd *cobra.Command, args []string) {
	if wf.IsRunning(convertJobName) {
		err := wf.Kill(convertJobName)
		if err != nil {
			logrus.Errorf("Error canceling job: %s", err)
		}
		alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json

		// remove tmp files
		os.RemoveAll(filepath.Join(alfred.GetOutput(wf), "_tmp"))

		// start notifier
		lib.Notifier("Cancel Background job")
	}
}

func init() {
	rootCmd.AddCommand(cancelCmd)
}
