/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>
*/
package cmd

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert youtub",
	Run:   runConvertCmd,
}

func runConvertCmd(ccmd *cobra.Command, args []string) {
	input, _ := ccmd.Flags().GetString("input")
	var url, title string
	buf := strings.Split(input, "\n")
	if len(buf) == 2 {
		url, title = buf[0], buf[1]
	} else if len(buf) == 1 {
		url = buf[0]
		title = buf[0]
	}

	// start notifier
	lib.Notifier("Start convert " + title)
	alfred.StoreOngoingProcess(wf, alfred.Process{
		Step:    "preparing",
		Process: "Preparing",
		Title:   title,
		URL:     url,
	})

	buf = strings.Split(args[0], ",")
	ss, t, fin, fout := buf[0], buf[1], buf[2], buf[3]

	srv, err := lib.NewConvert(lib.ConvertCfg{
		Url:     url,
		Title:   title,
		IsDebug: alfred.GetDebug(wf),
		Output:  alfred.GetOutput(wf),
		LoadOngoingProcess: func() (alfred.Process, error) {
			return alfred.LoadOngoingProcess(wf)
		},
		StoreOngoingProcess: func(prop alfred.Process) error {
			return alfred.StoreOngoingProcess(wf, prop)
		},
		LoadOngoingRingTone: func() (alfred.RingTone, error) {
			return alfred.LoadOngoingRingTone(wf)
		},
		StoreOngoingRingTone: func(prop alfred.RingTone) error {
			return alfred.StoreOngoingRingTone(wf, prop)
		},
	})
	if err != nil {
		srv.Reset()
		lib.Notifier(err.Error())
		logrus.Fatalf("failed to create new convert instance: %s", err.Error())
		return
	}

	// 1.download cover and get filename
	if err := srv.DownloadCoverAndGetFilename(); err != nil {
		srv.Reset()
		lib.Notifier(err.Error())
		return
	}

	// 2.download m4a
	if err := srv.DownloadM4a(ss, t); err != nil {
		srv.Reset()
		lib.Notifier(err.Error())
		return
	}

	// 3.identify target name and cover
	if err := srv.IdentifyTargetNameAndCover(); err != nil {
		lib.Notifier(err.Error())
		return
	}

	// 4.ffmpeg apply fadeIn fadeOut
	if err := srv.ApplyFadeInFadeOut(fin, fout); err != nil {
		srv.Reset()
		lib.Notifier(err.Error())
		return
	}

	// 5.add tag
	if err := srv.AddTag(); err != nil {
		srv.Reset()
		lib.Notifier(err.Error())
		return
	}

	// 6.reset
	targetName := srv.Reset()

	// start Notifier
	lib.Notifier(targetName + " is ready")
	lib.Detect(input)
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
