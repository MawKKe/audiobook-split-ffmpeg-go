package main

import (
	"flag"
	"fmt"
	"os"

	ffmpegsplit "github.com/MawKKe/audiobook-split-ffmpeg-go"
)

func main() {
	//flagVerbose := flag.Bool("verbose", false, "Show verbose (debug) output") // print stuff in main?
	flagInfile := flag.String("infile", "", "Input file path. REQUIRED.")
	flagOutdir := flag.String("outdir", "", "Output directory path. REQUIRED.")
	flagOnlyShowChaps := flag.Bool("only-show-chapters", false, "Only show parsed chapters, then exit. OPTIONAL")
	flagConcurrency := flag.Int("jobs", 0, "Number of concurrent ffmpeg jobs (default: num of cpus). OPTIONAL")
	//flagDryRun := flag.Bool("dry-run", false, "Only show which ffmpeg commands would run, without running them")
	//flagNoUseTitle := flag.Bool("no-use-title", false, "Only show which ffmpeg commands would run, without running them")

	flag.Parse()

	/*
	   if(*flagDryRun || *flagNoUseTitle){
	       panic("TODO flag -dry-run or NoUseTitle")
	   }
	*/

	// Both are required. However, the 'flag' package does not allow us
	// to specify that in the option declaration like python argparse does.
	if *flagInfile == "" || *flagOutdir == "" {
		flag.Usage()
		os.Exit(125)
	}

	imeta, err := ffmpegsplit.ReadFile(*flagInfile)

	if err != nil {
		fmt.Println(fmt.Errorf("Failed to read chapters: %v\n", err))
		os.Exit(1)
	} else if imeta.NumChapters() == 0 {
		fmt.Println("Error(?): Input file has no chapter metadata. Cannot continue.")
		os.Exit(1)
	}

	if *flagOnlyShowChaps {
		fmt.Printf("Found %v chapters:\n", imeta.NumChapters())
		for _, chap := range imeta.FFProbeOutput.Chapters {
			fmt.Printf("%+v\n", chap)
		}
		os.Exit(0)
	}

	opts := ffmpegsplit.DefaultOutFileOpts()

	workItems, err := imeta.ComputeWorkItems(*flagOutdir, opts)
	if err != nil {
		fmt.Printf("Failed to compute workitems: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Computed %v WorkItems\n", len(workItems))

	ffmpegsplit.Process(workItems, *flagConcurrency)
}
