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

// Package ffmpegsplit is for parsing chapter information from a multimedia file using FFProbe
package ffmpegsplit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ReadChaptersFromJSON parses the given byte sequence into a struct FFProbeOutput.
func ReadChaptersFromJSON(encoded []byte) (FFProbeOutput, error) {
	var decoded FFProbeOutput
	err := json.Unmarshal(encoded, &decoded)
	if err != nil {
		return FFProbeOutput{}, err
	}

	// find out what is the maximum chapter number. We don't assume that the
	// chapters are in any specific order.
	maxID := 0
	for _, chap := range decoded.Chapters {
		if chap.ID > maxID {
			maxID = chap.ID
		}
	}
	decoded.maxChapterID = maxID
	return decoded, nil
}

// ReadFile reads file metadata of file at path 'infile'
func ReadFile(infile string) (InputFileMetadata, error) {
	output, err := ReadChapters(infile)
	if err != nil {
		return InputFileMetadata{}, err
	}
	base := filepath.Base(infile)
	ext := filepath.Ext(base)
	basenoext := strings.TrimSuffix(base, ext)
	extnodot := strings.TrimPrefix(ext, ".")
	return InputFileMetadata{
		FFProbeOutput: output,
		Path:          infile,
		BaseNoExt:     basenoext,
		Extension:     extnodot,
	}, nil
}

// GetReadChaptersCommandline function builds the list of arguments used for
// reading chapter information via 'ffprobe' from file 'infile'.  Note: this
// function is called by ReadFile() - as such it is only useful for debug
// purposes.
func GetReadChaptersCommandline(infile string) []string {
	return []string{"-i", infile, "-v", "error", "-print_format", "json", "-show_chapters"}
}

// ReadChapters collects chapter information from the given file 'infile' using
// ffprobe. Blocks until subprocess returns. On success, parses the output
// (JSON) and returns the information in struct FFProbeOutput. Otherwise
// returns the error produced by either exec.Cmd.Run or json.Decoder.Unmarshal.
//
// Expects the program 'ffmpeg' to be somewhere in user's $PATH.
func ReadChapters(infile string) (FFProbeOutput, error) {
	args := GetReadChaptersCommandline(infile)
	cmd := exec.Command("ffprobe", args...)

	// capture output for further processing and/or error handling
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// NOTE: Runs in blocking mode
	err := cmd.Run()

	if err != nil {
		emsg := strings.TrimSuffix(stderr.String(), "\n")
		if emsg != "" {
			return FFProbeOutput{}, fmt.Errorf("ffprobe error: %s: %w", emsg, err)
		}
		return FFProbeOutput{}, fmt.Errorf("ffprobe error: %w", err)
	}

	return ReadChaptersFromJSON(stdout.Bytes())

}
