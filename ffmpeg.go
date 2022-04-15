// Copyright 2022 Markus Holmström (MawKKe)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
func computeOutname(outdir string, opts OutFileOpts, ch Chapter, imeta InputFileMetadata) string {
	baseName := imeta.BaseNoExt
	if Title, ok := ch.Tags["title"]; ok && opts.UseTitleInName {
		baseName = Title
	}

	// adjusted chapter Id
	num := ch.ID + opts.EnumOffset

	ext := imeta.Extension

	if opts.UseAlternateExtension != "" {
		ext = opts.UseAlternateExtension
	}

	return fmt.Sprintf("%0*d - %v.%v", opts.EnumPaddedWidth, num, baseName, ext)
}

// ComputeWorkItems processes struct workItem for each chapter. The workItem shall contain all
// the necessary information in order to extract the chapter using ffmpeg. When
// the sequence of workItems have been produced, the final processing step
// can be performed by calling workItem.Process().
func (imeta InputFileMetadata) ComputeWorkItems(outdir string, opts OutFileOpts) ([]WorkItem, error) {
	var wItems []WorkItem

	if opts.EnumOffset < 0 {
		opts.EnumOffset = 0
	}

	if opts.EnumPaddedWidth < 0 {
		maxChAdjusted := imeta.FFProbeOutput.maxChapterID + opts.EnumOffset
		opts.EnumPaddedWidth = len(fmt.Sprintf("%d", maxChAdjusted))
	}

	for _, chap := range imeta.FFProbeOutput.Chapters {
		outfile := computeOutname(outdir, opts, chap, imeta)
		wi := WorkItem{
			Infile:       imeta.Path,
			Outfile:      outfile,
			OutDirectory: outdir,
			Chapter:      chap,
			imeta:        imeta,
			opts:         opts,
		}
		wItems = append(wItems, wi)
	}

	return wItems, nil
}

// GetCommand produces a list of command line arguments that would produce the chapter file
// specific to this workItem
func (wi WorkItem) GetCommand() []string {
	return append([]string{"ffmpeg"}, wi.FFmpegArgs()...)
}

// FFmpegArgs converts a WorkItem to a list of arguments that are going to be passed to
// ffmpeg for actual processing step.
func (wi WorkItem) FFmpegArgs() []string {
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

	var metadataTrack []string
	if wi.opts.UseChapterNumberInMeta {
		off := wi.opts.EnumOffset
		metadataTrack = []string{"-metadata", fmt.Sprintf("track=%v/%v",
			int(wi.Chapter.ID)+off,
			wi.imeta.FFProbeOutput.maxChapterID+off)}
	}

	var metadataTitle []string

	if Title, ok := wi.Chapter.Tags["title"]; ok && Title != "" && wi.opts.UseTitleInMeta {
		metadataTitle = []string{"-metadata", fmt.Sprintf("title=%v", Title)}
	}

	args = append(args, metadataTrack...)
	args = append(args, metadataTitle...)
	args = append(args, filepath.Join(wi.OutDirectory, wi.Outfile))
	return args
}

// Process performs the actual processing step via ffmpeg.
// Expects 'ffmpeg' be somewhere in user's $PATH.
func (wi WorkItem) Process() error {
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
