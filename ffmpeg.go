package ffmpegsplit

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Chooses what the final chapter filename should be based on the options and
// available metadata.
func computeOutname(outdir string, opts OutFileOpts, ch *Chapter, imeta *InputFileMetadata) string {
	baseName := imeta.BaseNoExt
	if Title, ok := ch.Tags["title"]; ok && opts.UseTitleInName {
		baseName = Title
	}

	// adjusted chapter Id
	num := ch.Id + opts.EnumOffset

	return fmt.Sprintf("%0*d - %v.%v", opts.EnumPaddedWidth, num, baseName, imeta.Extension)
}

// Produces struct workItem for each chapter. The workItem shall contain all
// the necessary information in order to extract the chapter using ffmpeg. When
// the sequence of workItems have been produced, the final processing step
// can be performed by calling workItem.Process().
func (imeta *InputFileMetadata) ComputeWorkItems(outdir string, opts OutFileOpts) ([]workItem, error) {
	var w_items []workItem

	if opts.EnumOffset < 0 {
		opts.EnumOffset = 0
	}

	if opts.EnumPaddedWidth < 0 {
		maxChAdjusted := imeta.FFProbeOutput.maxChapterId + opts.EnumOffset
		opts.EnumPaddedWidth = len(fmt.Sprintf("%d", maxChAdjusted))
	}

	for i := range imeta.FFProbeOutput.Chapters {
		chap := &imeta.FFProbeOutput.Chapters[i]
		outfile := computeOutname(outdir, opts, chap, imeta)
		wi := workItem{
			Infile:       imeta.Path,
			Outfile:      outfile,
			OutDirectory: outdir,
			Chapter:      chap,
			imeta:        imeta,
			opts:         &opts,
		}
		w_items = append(w_items, wi)
	}

	return w_items, nil
}

// Produces a list of arguments that are going to be passed to FFMpeg for actual processing step.

func (wi *workItem) FFmpegArgs() []string {
	args := []string{
		"-nostdin",
		"-i", wi.imeta.Path,
		"-v", "error",
		"-map_chapters", "-1",
		"-vn",
		"-c", "copy",
		"-ss", wi.Chapter.StartTime,
		"-to", wi.Chapter.EndTime,
		"-n",
	}

	var metadata_track []string
	if wi.opts.UseChapterNumberInMeta {
		off := wi.opts.EnumOffset
		metadata_track = []string{"-metadata", fmt.Sprintf("track=%v/%v",
			int(wi.Chapter.Id)+off,
			wi.imeta.FFProbeOutput.maxChapterId+off)}
	}

	var metadata_title []string

	if Title, ok := wi.Chapter.Tags["title"]; ok && Title != "" && wi.opts.UseTitleInMeta {
		metadata_title = []string{"-metadata", fmt.Sprintf("title=%v", Title)}
	}

	args = append(args, metadata_track...)
	args = append(args, metadata_title...)
	args = append(args, filepath.Join(wi.OutDirectory, wi.Outfile))
	return args
}

// Performs the actual processing step via ffmpeg.
// Expects 'ffmpeg' be somewhere in user's $PATH.
func (wi *workItem) Process() error {
	const defaultPerm = 0755
	err := os.MkdirAll(wi.OutDirectory, defaultPerm)
	if err != nil {
		return err
	}

	// stdout should be empty on success
	// stderr will contain error message on failure
	var stderr bytes.Buffer
	cmd := exec.Command("ffmpeg", wi.FFmpegArgs()...)
	cmd.Stderr = &stderr

	// Blocks until completion
	err = cmd.Run()

	if err != nil {
		msg := strings.Trim(stderr.String(), "\n")
		return errors.New(msg)
	}

	return nil
}
