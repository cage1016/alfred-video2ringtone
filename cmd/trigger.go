/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const convertJobName = "go-convert-job"

// triggerCmd represents the trigger command
var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Trigger a convert job",
	Run:   runTriggerCmd,
}

func runTriggerCmd(cmd *cobra.Command, args []string) {
	input, _ := cmd.Flags().GetString("input")

	if !wf.IsRunning(convertJobName) {
		if len(args) != 1 {
			logrus.Fatalf("Expected 1 argument, got %d, %s", len(args), args)
		}

		ccmd := exec.Command(os.Args[0], "convert", "-i", input, "--", args[0])
		if err := wf.RunInBackground(convertJobName, ccmd); err != nil {
			logrus.Printf("Error starting update check: %s", err)
		}
	}
}

func init() {
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
