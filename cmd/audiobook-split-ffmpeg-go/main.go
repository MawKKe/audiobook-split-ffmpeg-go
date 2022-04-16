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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	ffmpegsplit "github.com/MawKKe/audiobook-split-ffmpeg-go"
	intervals "github.com/MawKKe/integer-interval-expressions-go"
)

func main() {
	//flagVerbose := flag.Bool("verbose", false, "Show verbose (debug) output") // print stuff in main?
	flagInfile := flag.String("infile", "", "Input file path. REQUIRED.")
	flagOutdir := flag.String("outdir", "", "Output directory path. REQUIRED.")
	flagOnlyShowChaps := flag.Bool("only-show-chapters", false, "Only show parsed chapters, then exit. OPTIONAL")
	flagOnlyShowCmds := flag.Bool("only-show-commands", false, "Only show final ffmpeg commands, then exit. OPTIONAL")
	flagConcurrency := flag.Int("jobs", 0, "Number of concurrent ffmpeg jobs (default: num of cpus). OPTIONAL")
	flagNoUseTitle := flag.Bool("no-use-title", false, "Only show which ffmpeg commands would run, without running them. OPTIONAL")
	flagSwapExt := flag.String("swap-extension", "", "Use this output file extension instead (WARNING: may force audio re-encoding)")
	flagSelectChapters := flag.String("select-chapters", "", "Only exctract these chapters")

	flag.Parse()

	// Both infile and outdir are required. However, the 'flag' package does not allow us
	// to specify that in the option declaration like python argparse does...
	var missing []string
	if *flagInfile == "" {
		missing = append(missing, "infile")
	}
	if *flagOutdir == "" {
		missing = append(missing, "outdir")
	}
	if len(missing) > 0 {
		fmt.Println(fmt.Errorf("ERROR: missing argument values %v", strings.Join(missing, ",")))
		flag.Usage()
		os.Exit(125)
	}

	imeta, err := ffmpegsplit.ReadFile(*flagInfile)

	if err != nil {
		fmt.Println(fmt.Errorf("Failed to read chapters: %w", err))
		os.Exit(1)
	} else if imeta.NumChapters() == 0 {
		fmt.Println("Error(?): Input file has no chapter metadata. Cannot continue.")
		os.Exit(2)
	}

	if *flagOnlyShowChaps {
		fmt.Printf("Found %v chapters:\n", imeta.NumChapters())
		for _, chap := range imeta.FFProbeOutput.Chapters {
			fmt.Printf("%+v\n", chap)
		}
		os.Exit(0)
	}

	opts := ffmpegsplit.DefaultOutFileOpts()

	opts.UseTitleInName = !*flagNoUseTitle
	opts.UseAlternateExtension = *flagSwapExt

	if *flagSelectChapters != "" {
		expression, err := intervals.ParseExpression(*flagSelectChapters)
		if err != nil {
			fmt.Println("error in select-chapters:", err)
			os.Exit(2)
		}
		selectChapter := func(ch ffmpegsplit.Chapter) bool {
			return !expression.Matches(ch.ID)
		}

		opts.AddFilter(ffmpegsplit.ChapterFilter{
			Description: "Select by chapter ID", Filter: selectChapter,
		})
	}

	workItems, err := imeta.ComputeWorkItems(*flagOutdir, opts)
	if err != nil {
		fmt.Printf("Failed to compute workitems: %v\n", err)
		os.Exit(2)
	}

	//fmt.Printf("Computed %v WorkItems\n", len(workItems))

	if *flagOnlyShowCmds {
		for i := range workItems {
			fmt.Println(strings.Join(escapeCmd(workItems[i].GetCommand()), " "))
		}
		os.Exit(0)
	}

	status := ffmpegsplit.Process(workItems, *flagConcurrency)
	fmt.Println("Status:", status)
}

func escapeCmd(unescaped []string) []string {
	var escaped []string
	for _, s := range unescaped {
		escaped = append(escaped, strconv.Quote(s))
	}
	return escaped
}
