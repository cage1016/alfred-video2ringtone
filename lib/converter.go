package lib

import (
	"bufio"
	"io"
	"regexp"
)

type Converter interface {
	DownloadCoverAndGetFilename() error
	DownloadM4a(ss, t string) error
	IdentifyTargetNameAndCover() error
	ApplyFadeInFadeOut(fin, fout string) error
	AddTag() error
	Reset() string
}

var thumbnailRegex = regexp.MustCompile(`(?m)to:\s*(.*)$`)
var m4aRegex = regexp.MustCompile(`(?m)Destination:\s*(.*)$`)

func worker(done chan struct{}, r io.ReadCloser, fn func(string)) {
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			fn(scanner.Text())
		}
		done <- struct{}{}
	}()
}

func lastItem(ss []string) string {
	return ss[len(ss)-1]
}
