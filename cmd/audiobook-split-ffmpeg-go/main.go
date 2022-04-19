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

type ProgramArgs struct {
	InFile          string
	OutDir          string
	OnlyShowChaps   bool
	OnlyShowCmds    bool
	Concurrency     int
	NoUseTitle      bool
	SwapExt         string
	SelectByChapter ffmpegsplit.ChapterFilterFunction
}

func ParseCommandline() (args ProgramArgs) {
	flag.StringVar(&args.InFile, "infile", "",
		"Input file path. REQUIRED.")
	flag.StringVar(&args.OutDir, "outdir", "",
		"Output directory path. REQUIRED.")
	flag.BoolVar(&args.OnlyShowChaps, "only-show-chapters", false,
		"Only show parsed chapters, then exit. OPTIONAL")
	flag.BoolVar(&args.OnlyShowCmds, "only-show-commands", false,
		"Only show final ffmpeg commands, then exit. OPTIONAL")
	flag.IntVar(&args.Concurrency, "jobs", 0,
		"Number of concurrent ffmpeg jobs (default: num of cpus). OPTIONAL")
	flag.BoolVar(&args.NoUseTitle, "no-use-title", false,
		"Only show which ffmpeg commands would run, without running them. OPTIONAL")
	flag.StringVar(&args.SwapExt, "swap-extension", "",
		"Use this output file extension instead (WARNING: may force audio re-encoding)")

	var selectChaptersHelp string = "Exctract only the specified chapters.\n" +
		"The argument value should be a comma-separated list of chapter\n" +
		"numbers or ranges of chapter numbers. For example '1,3-5,7-'"

	flag.Func("select-chapters", selectChaptersHelp, func(exprStr string) error {
		if exprStr == "" {
			return fmt.Errorf("expression would match 0 chapters")
		}
		if exprStr == "*" {
			// match all chapters, default behavior
			return nil
		}
		expression, err := intervals.ParseExpression(exprStr)
		if err != nil {
			return err
		}
		args.SelectByChapter = func(ch ffmpegsplit.Chapter) bool {
			return !expression.Matches(ch.ID)
		}
		return nil
	})

	flag.Parse()

	// Both infile and outdir are required. However, the 'flag' package does not allow us
	// to specify that in the option declaration like python argparse does...
	var missing []string
	if args.InFile == "" {
		missing = append(missing, "infile")
	}
	if args.OutDir == "" {
		missing = append(missing, "outdir")
	}
	if len(missing) > 0 {
		fmt.Println(fmt.Errorf("ERROR: missing argument values %v", strings.Join(missing, ",")))
		flag.Usage()
		os.Exit(125)
	}

	return
}

func main() {
	args := ParseCommandline()

	imeta, err := ffmpegsplit.ReadFile(args.InFile)

	if err != nil {
		fmt.Println(fmt.Errorf("Failed to read chapters: %w", err))
		os.Exit(1)
	} else if imeta.NumChapters() == 0 {
		fmt.Println("Error(?): Input file has no chapter metadata. Cannot continue.")
		os.Exit(2)
	}

	if args.OnlyShowChaps {
		fmt.Printf("Found %v chapters:\n", imeta.NumChapters())
		for _, chap := range imeta.FFProbeOutput.Chapters {
			fmt.Printf("%+v\n", chap)
		}
		os.Exit(0)
	}

	opts := ffmpegsplit.DefaultOutFileOpts()

	opts.UseTitleInName = !args.NoUseTitle
	opts.UseAlternateExtension = args.SwapExt

	if args.SelectByChapter != nil {
		opts.AddFilter(ffmpegsplit.ChapterFilter{
			Description: "Select by chapter ID", Filter: args.SelectByChapter,
		})
	}

	workItems, err := imeta.ComputeWorkItems(args.OutDir, opts)
	if err != nil {
		fmt.Printf("Failed to compute workitems: %v\n", err)
		os.Exit(2)
	}

	//fmt.Printf("Computed %v WorkItems\n", len(workItems))

	if args.OnlyShowCmds {
		for i := range workItems {
			fmt.Println(strings.Join(escapeCmd(workItems[i].GetCommand()), " "))
		}
		os.Exit(0)
	}

	status := ffmpegsplit.Process(workItems, args.Concurrency)
	fmt.Println("Status:", status)
}

func escapeCmd(unescaped []string) []string {
	var escaped []string
	for _, s := range unescaped {
		escaped = append(escaped, strconv.Quote(s))
	}
	return escaped
}
