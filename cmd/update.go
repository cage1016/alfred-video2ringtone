/*
Copyright © 2023 KAI CHU CHUNG
*/
package cmd

import (
	"log"
	"os"
	"os/exec"

	aw "github.com/deanishe/awgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	EnabledAutoUpdate = "true"
)

const updateJobName = "checkForUpdate"

func CheckForUpdate() {
	if EnabledAutoUpdate == "true" {
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
	} else {
		logrus.Debug("Auto update disabled")
	}
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Youtube 2 Ringtone",
	Run: func(cmd *cobra.Command, args []string) {
		if EnabledAutoUpdate == "true" {
			wf.Configure(aw.TextErrors(true))
			log.Println("Checking for updates...")
			if err := wf.CheckForUpdate(); err != nil {
				wf.FatalError(err)
			}
		} else {
			logrus.Debug("Auto update disabled")
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
