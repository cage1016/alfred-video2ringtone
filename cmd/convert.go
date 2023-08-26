/*
Copyright © 2022 KAI CHU CHUNG <cage.chung@gmail.com>

*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cage1016/alfred-yt2ringtone/alfred"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const p = `tell application id "com.runningwithcrayons.Alfred" to run trigger "notifier" in workflow "com.kaichu.yt2ringtone" with argument "%s"`

func lastItem(ss []string) string {
	return ss[len(ss)-1]
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func notifier(title, msg string) {
	exec.Command("osascript", "-e", fmt.Sprintf(p, strings.Join([]string{title, msg}, ","))).Run()
}

func worker(done chan struct{}, r io.ReadCloser, fn func(string)) {
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			fn(scanner.Text())
		}
		done <- struct{}{}
	}()
}

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
	notifier("YouTube 2 Ringtone", "Start convert "+title)
	alfred.StoreOngoingProcess(wf, alfred.Process{
		Step:    "preparing",
		Process: "Preparing",
		Title:   title,
		URL:     url,
	})

	buf = strings.Split(args[0], ",")
	ss, t, fin, fout := buf[0], buf[1], buf[2], buf[3]
	var targetName string
	{
		flags := []string{
			url,
			"--external-downloader", "ffmpeg",
			"--external-downloader-args", fmt.Sprintf("ffmpeg_i:-ss %s -t %s", ss, t),
			"-x", "--audio-format", "m4a",
			"--yes-overwrites",
			"--ignore-errors",
			"--output", filepath.Join(wf.DataDir(), "%(title)s_tmp.%(ext)s"),
		}

		// get file name
		{
			cmd := exec.Command("yt-dlp", append([]string{"--get-filename"}, flags...)...)
			logrus.Debugf("yt-dlp get-filename: %s", cmd)
			out, _ := cmd.CombinedOutput()
			logrus.Debugf("yt-dlp get-filename output: %s", out)
			if strings.HasPrefix(string(out), "ERROR") {
				notifier("YouTube 2 Ringtone", string(out))
				alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json
				logrus.Fatalf("yt-dlp error: %s", string(out))
			}
			targetName = fileNameWithoutExtTrimSuffix(filepath.Base(strings.Split(string(out), "\n")[0]))
		}

		// download
		{
			cmd := exec.Command("yt-dlp", append([]string{"--newline", "-f", "bestaudio[ext=m4a]"}, flags...)...)
			logrus.Debugf("yt-dlp download: %s", cmd)
			r, _ := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout

			done := make(chan struct{})
			worker(done, r, func(line string) {
				logrus.Debug(line)
				data, _ := alfred.LoadOngoingProcess(wf)
				data.Process = line
				data.Step = "downloading"
				alfred.StoreOngoingProcess(wf, data)
			})

			cmd.Start()
			<-done
			cmd.Wait()

			// check if download success
			data, _ := alfred.LoadOngoingProcess(wf)
			if strings.HasPrefix(data.Process, "ERROR") {
				notifier("YouTube 2 Ringtone", data.Process)
				alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json
				logrus.Fatalf("yt-dlp error: %s", data.Process)
			}
		}
	}

	DestName := strings.TrimSuffix(targetName, "_tmp")
	{
		// ffmpeg apply fadeIn fadeOut
		flags := []string{
			"-y",
			"-i", filepath.Join(wf.DataDir(), targetName+".m4a"),
			"-filter_complex", fmt.Sprintf("afade=d=%s, areverse, afade=d=%s, areverse", fin, fout),
			filepath.Join(wf.DataDir(), DestName+".m4a"),
		}
		cmd := exec.Command("ffmpeg", flags...)
		logrus.Debugf("ffmpeg apply fadeIn fadeOut cmd: %s", cmd)
		r, _ := cmd.StdoutPipe()
		cmd.Stderr = cmd.Stdout

		done := make(chan struct{})
		worker(done, r, func(line string) {
			logrus.Debug(line)
			if strings.HasPrefix(line, "size=") {
				data, _ := alfred.LoadOngoingProcess(wf)
				data.Process = strings.TrimSpace(lastItem(strings.Split(line, "\r")))
				data.Step = "trimming"
				alfred.StoreOngoingProcess(wf, data)
			}
		})

		cmd.Start()
		<-done
		cmd.Wait()
	}

	// reset
	alfred.StoreOngoingProcess(wf, alfred.Process{})
	data, _ := alfred.LoadOngoingRingTone(wf)
	if data.Item == nil {
		data.Item = map[string]alfred.M4a{}
	}
	data.Item[url] = alfred.M4a{
		Name:      DestName + ".m4a",
		Info:      fmt.Sprintf("Start %s with %ss duration, %ss fadeIn, %ss fadeOut", ss, t, fin, fout),
		CreatedAt: time.Now().Unix(),
	}
	alfred.StoreOngoingRingTone(wf, data)

	// remove temporary file
	os.Remove(filepath.Join(wf.DataDir(), targetName+".m4a"))

	// start notifier
	notifier("YouTube 2 Ringtone", DestName+".m4a is ready")
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
