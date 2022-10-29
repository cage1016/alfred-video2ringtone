/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>

*/
package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cage1016/alfred-yt2ringtone/alfred"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/spf13/cobra"
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

func CheckForUpdate() {
	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")
		cmd := exec.Command(os.Args[0], "update")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	if wf.UpdateAvailable() {
		wf.Configure(aw.SuppressUIDs(true))
		wf.NewItem("An update is available!").
			Subtitle("⇥ or ↩ to install update").
			Valid(false).
			Autocomplete("workflow:update").
			Icon(&aw.Icon{Value: "update-available.png"})
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "Ringtone Maker via Youtube",
	Short: "Create ringtone from youtube video",
	Run: func(cmd *cobra.Command, args []string) {
		CheckForUpdate()
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
			log.Println(err)
			os.Exit(1)
		}
	})
}

func init() {
	wf = aw.New(update.GitHub(repo), aw.HelpURL(repo+"/issues"))
	wf.Args() // magic for "workflow:update"
}
