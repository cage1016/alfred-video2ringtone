package lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wellmoon/m4aTag/mtag"

	"github.com/cage1016/alfred-video2ringtone/alfred"
)

type ConvertCfg struct {
	Output  string
	IsDebug bool

	Url   string
	Title string

	LoadOngoingProcess   func() (alfred.Process, error)
	StoreOngoingProcess  func(prop alfred.Process) error
	LoadOngoingRingTone  func() (alfred.RingTone, error)
	StoreOngoingRingTone func(prop alfred.RingTone) error
}

type Convert struct {
	isDebug bool           // debug mode
	log     *logrus.Logger // logger
	output  string         // output dir
	url     string         // video url
	title   string         // video title

	covers      []string
	targetNames []string

	cover            string
	targetName       string
	ss, t, fin, fout string

	loadOngoingProcess   func() (alfred.Process, error)
	storeOngoingProcess  func(prop alfred.Process) error
	loadOngoingRingTone  func() (alfred.RingTone, error)
	storeOngoingRingTone func(prop alfred.RingTone) error
}

// DownloadCoverAndGetFilename implements Converter.
func (c *Convert) DownloadCoverAndGetFilename() error {
	c.log.Debugf("1. yt-dlp download cover and get filename: %s", c.url)

	flags := []string{
		"--skip-download",
		"--write-thumbnail",
		"-o",
		filepath.Join(c.output, "%(title)s.%(ext)s"),
		c.url,
	}
	cmd := exec.Command("yt-dlp", flags...)
	c.log.Debugf("yt-dlp download cover and get filename: %s", cmd)

	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	done := make(chan struct{})
	worker(done, r, func(line string) {
		c.log.Debug(line)
		data, _ := c.loadOngoingProcess()
		data.Process = line
		data.Step = "downloading-cover"
		c.storeOngoingProcess(data)

		for _, match := range thumbnailRegex.FindAllString(line, -1) {
			c.covers = append(c.covers, strings.TrimPrefix(match, "to: "))
		}
	})

	cmd.Start()
	<-done
	cmd.Wait()

	// check if download success
	data, _ := c.loadOngoingProcess()
	if strings.HasPrefix(data.Process, "ERROR") {
		c.log.Errorf("yt-dlp error: %s", data.Process)
		return fmt.Errorf("yt-dlp error: %s", data.Process)
	}

	return nil
}

// DownloadM4a implements Converter.
func (c *Convert) DownloadM4a(ss, t string) error {
	c.log.Debugf("2. yt-dlp download m4a with ss, t : %s", c.url, ss, t)

	c.ss = ss
	c.t = t

	flags := []string{
		c.url,
		"--external-downloader", "ffmpeg",
		"--external-downloader-args", fmt.Sprintf("ffmpeg_i:-ss %s -t %s", ss, t),
		"-x", "--audio-format", "m4a",
		"--yes-overwrites", "--ignore-errors",
		"--output", filepath.Join(c.output, "%(title)s.%(ext)s"),
	}

	cmd := exec.Command("yt-dlp", append([]string{"--newline", "-f", "bestaudio"}, flags...)...)
	c.log.Debugf("yt-dlp download: %s", cmd)

	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	done := make(chan struct{})
	worker(done, r, func(line string) {
		c.log.Debug(line)
		data, _ := c.loadOngoingProcess()
		data.Process = line
		data.Step = "downloading"
		c.storeOngoingProcess(data)

		for _, match := range m4aRegex.FindAllString(line, -1) {
			c.targetNames = append(c.targetNames, strings.TrimPrefix(match, "Destination: "))
		}
	})

	cmd.Start()
	<-done
	cmd.Wait()

	// check if download success
	data, _ := c.loadOngoingProcess()
	if strings.HasPrefix(data.Process, "ERROR") {
		c.log.Errorf("yt-dlp error: %s", data.Process)
		return fmt.Errorf("yt-dlp error: %s", data.Process)
	}

	return nil
}

// IdentifyTargetNameAndCover() error
func (c *Convert) IdentifyTargetNameAndCover() error {
	c.log.Debugf("3. identify target name and cover: %s", c.url)

	c.targetName = lastItem(c.targetNames)
	c.cover = lastItem(c.covers)
	if strings.HasSuffix(c.cover, ".webp") {
		output := filepath.Join(c.output, strings.TrimSuffix(filepath.Base(c.cover), ".webp")+".jpg")
		cmd := exec.Command("ffmpeg", "-y", "-i", c.cover, output)
		err := cmd.Run()
		if err != nil {
			c.log.Errorf("ffmpeg convert webp to jpg error: %s", err)
			return fmt.Errorf("ffmpeg convert webp to jpg error: %s", err)
		}
		c.cover = output
	}
	return nil
}

// ApplyFadeInFadeOut implements Converter.
func (c *Convert) ApplyFadeInFadeOut(fin, fout string) error {
	c.log.Debugf("4. apply %s fadeIn %s fadeOut: %s", c.url, fin, fout)

	c.fin = fin
	c.fout = fout

	flags := []string{
		"-y",
		"-i", c.targetName,
		"-filter_complex", fmt.Sprintf("afade=d=%s, areverse, afade=d=%s, areverse", fin, fout),
		c.targetName,
	}
	cmd := exec.Command("ffmpeg", flags...)
	c.log.Debugf("ffmpeg apply fadeIn fadeOut cmd: %s", cmd)
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	done := make(chan struct{})
	worker(done, r, func(line string) {
		c.log.Debug(line)
		if strings.HasPrefix(line, "size=") {
			data, _ := c.loadOngoingProcess()
			data.Process = strings.TrimSpace(lastItem(strings.Split(line, "\r")))
			data.Step = "trimming"
			c.storeOngoingProcess(data)
		}
	})

	cmd.Start()
	<-done
	cmd.Wait()

	return nil
}

// AddTag implements Converter.
func (c *Convert) AddTag() error {
	c.log.Debugf("5. update m4a tag: %s", c.url)

	err := mtag.UpdateM4aTag(true,
		c.targetName,
		strings.TrimSuffix(filepath.Base(c.targetName), filepath.Ext(c.targetName)),
		"",
		"",
		c.url,
		c.cover,
	)
	if err != nil {
		c.log.Errorf("update m4a tag error: %s", err)
		return fmt.Errorf("update m4a tag error: %s", err)
	}
	return nil
}

// Reset implements Converter.
func (c *Convert) Reset() string {
	c.log.Debugf("6. reset ongoing process: %s", c.url)

	// reset ongoing process
	c.storeOngoingProcess(alfred.Process{})

	// update ongoing ringtone
	{
		data, _ := c.loadOngoingRingTone()
		if data.Items == nil {
			data.Items = map[string]alfred.M4a{}
		}
		data.Items[c.url] = alfred.M4a{
			Title:     c.title,
			Name:      filepath.Base(c.targetName),
			Info:      fmt.Sprintf("Start %s with %ss duration, %ss fadeIn, %ss fadeOut", c.ss, c.t, c.fin, c.fout),
			CreatedAt: time.Now().Unix(),
		}
		c.storeOngoingRingTone(data)
	}

	// move c.targetName to parent directory
	newFilePath := filepath.Join(filepath.Dir(c.targetName), "..", filepath.Base(c.targetName))
	os.Rename(c.targetName, newFilePath)

	// delete temporary directory
	os.RemoveAll(c.output)

	return newFilePath
}

// initLogger writes logs to STDOUT and a.paths.DAGDir/wallet.log
func (c *Convert) initLogger() {
	logf := c.output + "/convert.log"
	if c.isDebug {
		c.log.SetLevel(logrus.DebugLevel)

		logFile, err := os.OpenFile(logf, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			c.log.Fatalf("open log file failed: %v", err)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		c.log.SetOutput(mw)
		c.log.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
	} else {
		if _, err := os.Stat(logf); err == nil {
			os.Remove(logf)
		}
	}
}

// create directories
func (c *Convert) directoryCreator(directories ...string) error {
	for _, d := range directories {
		err := os.MkdirAll(d, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewConvert(cfg ConvertCfg) (Converter, error) {
	srv := &Convert{
		log:                  logrus.New(),
		output:               cfg.Output,
		isDebug:              cfg.IsDebug,
		url:                  cfg.Url,
		title:                cfg.Title,
		covers:               []string{},
		targetNames:          []string{},
		loadOngoingProcess:   cfg.LoadOngoingProcess,
		storeOngoingProcess:  cfg.StoreOngoingProcess,
		loadOngoingRingTone:  cfg.LoadOngoingRingTone,
		storeOngoingRingTone: cfg.StoreOngoingRingTone,
	}

	// init logger
	srv.initLogger()

	// wrap output dir
	srv.output = filepath.Join(srv.output, "_tmp")
	err := srv.directoryCreator(srv.output)
	if err != nil {
		return nil, errors.Wrap(err, "create directory failed")
	}

	return srv, nil
}
