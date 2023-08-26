/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"errors"
	"os"
	"path/filepath"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-yt2ringtone/alfred"
)

const updateJobName = "checkForUpdate"

var (
	repo = "cage1016/alfred-yt-ringtone"
	wf   *aw.Workflow
	av   = aw.NewArgVars()
)

func ErrorHandle(err error) {
	av.Var("error", err.Error())
	if err := av.Send(); err != nil {
		wf.Fatalf("failed to send args to Alfred: %v", err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Ringtone Maker via Youtube",
	Short: "Create ringtone from youtube video",
	Run: func(cmd *cobra.Command, args []string) {
		wf.SendFeedback()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	wf.Run(func() {
		if _, err := os.Stat(filepath.Join(wf.DataDir(), "process.json")); errors.Is(err, os.ErrNotExist) {
			alfred.StoreOngoingProcess(wf, alfred.Process{})
		}

		if _, err := os.Stat(filepath.Join(wf.DataDir(), "ringtone.json")); errors.Is(err, os.ErrNotExist) {
			alfred.StoreOngoingRingTone(wf, alfred.RingTone{})
		}

		if err := rootCmd.Execute(); err != nil {
			logrus.Fatal(err)
		}
	})
}

func init() {
	wf = aw.New(update.GitHub(repo), aw.HelpURL(repo+"/issues"))
	wf.Args() // magic for "workflow:update"

	if alfred.GetDebug(wf) {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}
