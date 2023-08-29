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

var thumbnailRegex = regexp.MustCompile(`(?m)to:\s*(.*)$`)
var m4aRegex = regexp.MustCompile(`(?m)Destination:\s*(.*)$`)

func lastItem(ss []string) string {
	return ss[len(ss)-1]
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
	var covers []string
	var targetNames []string

	// download cover and get filename
	{
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

			for _, match := range thumbnailRegex.FindAllString(line, -1) {
				covers = append(covers, strings.TrimPrefix(match, "to: "))
			}
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

	// download m4a
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

			for _, match := range m4aRegex.FindAllString(line, -1) {
				targetNames = append(targetNames, strings.TrimPrefix(match, "Destination: "))
			}
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

	targetName := lastItem(targetNames)
	cover := lastItem(covers)
	if strings.HasSuffix(cover, ".webp") {
		output := filepath.Join(alfred.GetOutput(wf), strings.TrimSuffix(filepath.Base(cover), ".webp")+".jpg")
		cmd := exec.Command("ffmpeg", "-i", cover, output)
		err := cmd.Run()
		if err != nil {
			logrus.Errorf("ffmpeg convert webp to jpg error: %s", err)
		}
		cover = output
	}

	// ffmpeg apply fadeIn fadeOut
	{
		flags := []string{
			"-y",
			"-i", targetName,
			"-filter_complex", fmt.Sprintf("afade=d=%s, areverse, afade=d=%s, areverse", fin, fout),
			targetName,
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
		err := mtag.UpdateM4aTag(true,
			targetName,
			strings.TrimSuffix(filepath.Base(targetName), filepath.Ext(targetName)),
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
		Name:      filepath.Base(targetName),
		Info:      fmt.Sprintf("Start %s with %ss duration, %ss fadeIn, %ss fadeOut", ss, t, fin, fout),
		CreatedAt: time.Now().Unix(),
	}
	alfred.StoreOngoingRingTone(wf, data)

	// remove temporary cover file and m4a file
	for _, cover := range covers {
		os.Remove(cover)
	}

	for _, t := range targetNames {
		if t != targetName {
			os.Remove(t)
		}
	}

	// start Notifier
	lib.Notifier(targetName + " is ready")
	lib.Detect(input)
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.PersistentFlags().StringP("input", "i", "", "youtube url")
}
