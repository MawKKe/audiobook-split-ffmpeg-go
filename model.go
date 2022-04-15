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

// Chapter represents a single chapter in ffprobe output JSON
type Chapter struct {
	ID        int               `json:"id"`
	TimeBase  string            `json:"time_base"` // float or fixnum? Not needed anyways
	Start     int               `json:"start"`
	StartTime string            `json:"start_time"` // float or fixnum? Not needed anyways
	End       int               `json:"end"`
	EndTime   string            `json:"end_time"` // float or fixnum? Not needed anyways
	Tags      map[string]string `json:"tags"`
}

// FFProbeOutput represents the JSON structure returned by ffprobe command
type FFProbeOutput struct {
	Chapters     []Chapter `json:"chapters"`
	maxChapterID int       // hacky, but works..?
}

// InputFileMetadata tepresents all important details of the input file.
// Produced by ReadFile().
type InputFileMetadata struct {
	Path          string
	BaseNoExt     string
	Extension     string
	FFProbeOutput FFProbeOutput
}

// NumChapters returns the number of chapters found in the input file.
func (imeta InputFileMetadata) NumChapters() int {
	return len(imeta.FFProbeOutput.Chapters)
}

// WorkItem represents all the required information for processing the input
// file into a chapter specific file. To do the actual processing,
// run WorkItem.Process()
type WorkItem struct {
	Infile       string
	Outfile      string
	OutDirectory string
	Chapter      Chapter
	imeta        InputFileMetadata
	opts         OutFileOpts
}

// OutFileOpts contains user-defined options specifying how the output files
// will be named and what kind of metadata they shall contain (if metadata even
// is available in the original input file).
type OutFileOpts struct {
	// Place chapter title in output file name? (NOTE: Only if title is available)
	UseTitleInName bool

	// Place chapter title in output file metadata? (NOTE: Only if title is available)
	UseTitleInMeta bool

	// Place chapter number in output file metadata?
	UseChapterNumberInMeta bool

	// Adjusts the starting value of filename enumeration. Sometimes it
	// might make more sense to start enumeration from 1 instead of 0, for example.
	// Negative value tells the library to choose automatically.
	EnumOffset int

	// When chapter number is used in the filename, the number may be
	// left-padded with zeros in order to produce constant-width "column" of chapter numbers.
	// This has the advantage that files can now be sorted more easily by various *nix tools.
	//
	// This flag specifies how many leading zeros should in the filename enumeration, if at all.
	// Set value to <0 to let the library automatically compute the appropriate padding.
	// Set valut to  0 to disable padding
	// Otherwise, the value will determine the number of leading zeros.
	EnumPaddedWidth int
}

// DefaultOutFileOpts returns some sensible set of default values for OutFileOpts.
func DefaultOutFileOpts() OutFileOpts {
	var opts OutFileOpts
	opts.UseTitleInName = true
	opts.UseTitleInMeta = true
	opts.UseChapterNumberInMeta = true
	opts.EnumOffset = -1
	opts.EnumPaddedWidth = -1
	return opts

}
