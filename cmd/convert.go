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
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wellmoon/m4aTag/mtag"

	"github.com/cage1016/alfred-video2ringtone/alfred"
	"github.com/cage1016/alfred-video2ringtone/lib"
)

func lastItem(ss []string) string {
	return ss[len(ss)-1]
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
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
	lib.Notifier("Start convert " + title)
	alfred.StoreOngoingProcess(wf, alfred.Process{
		Step:    "preparing",
		Process: "Preparing",
		Title:   title,
		URL:     url,
	})

	buf = strings.Split(args[0], ",")
	ss, t, fin, fout := buf[0], buf[1], buf[2], buf[3]
	var targetName, cover string
	{

		// download cover and get filename
		{
			// cmd := exec.Command("yt-dlp", append([]string{"--get-filename"}, flags...)...)
			// logrus.Debugf("yt-dlp get-filename: %s", cmd)
			// out, _ := cmd.CombinedOutput()
			// logrus.Debugf("yt-dlp get-filename output: %s", out)
			// if strings.HasPrefix(string(out), "ERROR") {
			// 	lib.Notifier(string(out))
			// 	alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json
			// 	logrus.Fatalf("yt-dlp error: %s", string(out))
			// }
			// cover = strings.Split(string(out), "\n")[0]
			// targetName = fileNameWithoutExtTrimSuffix(filepath.Base(cover))

			flags := []string{
				"--skip-download", "--write-thumbnail",
				"-o", filepath.Join(alfred.GetOutput(wf), "%(title)s.%(ext)s"),
				url,
			}
			cmd := exec.Command("yt-dlp", flags...)
			logrus.Debugf("yt-dlp download cover and get filename: %s", cmd)

			r, _ := cmd.StdoutPipe()
			cmd.Stderr = cmd.Stdout

			done := make(chan struct{})
			worker(done, r, func(line string) {
				logrus.Debug(line)
				data, _ := alfred.LoadOngoingProcess(wf)
				data.Process = line
				data.Step = "downloading-cover"
				alfred.StoreOngoingProcess(wf, data)
			})

			cmd.Start()
			<-done
			cmd.Wait()

			// check if download success
			data, _ := alfred.LoadOngoingProcess(wf)
			if strings.HasPrefix(data.Process, "ERROR") {
				lib.Notifier(data.Process)
				alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json
				logrus.Fatalf("yt-dlp error: %s", data.Process)
			}

			var re = regexp.MustCompile(`(?m)(/[^"]+\.*)`)
			x := re.FindAllString(data.Process, -1)
			cover = x[0]
			targetName = fileNameWithoutExtTrimSuffix(filepath.Base(cover))
		}

		// download
		{
			flags := []string{
				url,
				"--external-downloader", "ffmpeg",
				"--external-downloader-args", fmt.Sprintf("ffmpeg_i:-ss %s -t %s", ss, t),
				"-x", "--audio-format", "m4a",
				"--yes-overwrites",
				"--ignore-errors",
				"--output", filepath.Join(alfred.GetOutput(wf), "%(title)s.%(ext)s"),
			}

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
				lib.Notifier(data.Process)
				alfred.StoreOngoingProcess(wf, alfred.Process{}) // reset process.json
				logrus.Fatalf("yt-dlp error: %s", data.Process)
			}
		}
	}

	// DestName := strings.TrimSuffix(targetName, "_tmp")
	{
		// ffmpeg apply fadeIn fadeOut
		flags := []string{
			"-y",
			"-i", filepath.Join(alfred.GetOutput(wf), targetName+".m4a"),
			"-filter_complex", fmt.Sprintf("afade=d=%s, areverse, afade=d=%s, areverse", fin, fout),
			filepath.Join(alfred.GetOutput(wf), targetName+".m4a"),
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

	// add tag
	{
		if strings.HasSuffix(cover, ".webp") {
			output := filepath.Join(alfred.GetOutput(wf), targetName+".jpg")
			cmd := exec.Command("ffmpeg", "-i", cover, output)
			err := cmd.Run()
			if err != nil {
				logrus.Errorf("ffmpeg convert webp to jpg error: %s", err)
			}
			os.Remove(cover)
			cover = output
		}

		err := mtag.UpdateM4aTag(true,
			filepath.Join(alfred.GetOutput(wf), targetName+".m4a"),
			targetName,
			"",
			"",
			url,
			cover,
		)
		if err != nil {
			logrus.Errorf("update m4a tag error: %s", err)
		}
	}

	// reset
	alfred.StoreOngoingProcess(wf, alfred.Process{})
	data, _ := alfred.LoadOngoingRingTone(wf)
	if data.Items == nil {
		data.Items = map[string]alfred.M4a{}
	}
	data.Items[url] = alfred.M4a{
		Title:     title,
		Name:      targetName + ".m4a",
		Info:      fmt.Sprintf("Start %s with %ss duration, %ss fadeIn, %ss fadeOut", ss, t, fin, fout),
		CreatedAt: time.Now().Unix(),
	}
	alfred.StoreOngoingRingTone(wf, data)

	// remove temporary cover file
	os.Remove(cover)

	// start Notifier
	lib.Notifier(targetName + ".m4a is ready")
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
