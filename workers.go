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
	"fmt"
	"runtime"
	"sync"
)

type job struct {
	wi *workItem
}
type result struct {
	wi  *workItem
	err error
}

// Describes how many chapter extractions succeeded and how many failed.
// Note that successful + failed should equal submitted, otherwise an error
// happened somewhere.
type Status struct {
	Successful int
	Failed     int
	Submitted  int
}

// Process all workItems, i.e. do the actual extraction process. The workItems
// contain all the necessary information for the extractions to be performed.
// The processing happens in parallel, using at most 'maxConcurrent' ffmpeg
// worker processes.
//
// Note: the extraction process does not re-encode the audio stream, thus the
// processing performance is not likely CPU-bound. However, using too many
// workers extracting the same file may saturate I/O, decreasing overall
// performance. In summary: increasing 'maxConcurrent' value may improve
// performance, but only up to a point.
//
// TODO: add similar processin interface with support for context.Context (use
// exec.CommandContext?)
func Process(workItems []workItem, maxConcurrent int) Status {
	if maxConcurrent <= 0 {
		maxConcurrent = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	chJob := make(chan job, len(workItems))
	chRes := make(chan result, len(workItems))
	for t := 0; t < maxConcurrent; t++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range chJob {
				err := job.wi.Process()
				chRes <- result{job.wi, err}
			}
		}()
	}
	// the channel was created with enough room to hold all jobs,
	// so this should finish immediately
	for i := range workItems {
		chJob <- job{&workItems[i]}
	}

	var failed int
	var successful int
	for i := 0; i < len(workItems); i++ {
		// TODO This receive may block indefinetely. Use select with timeouts?
		// ALTHOUGH a large chapter may take a long time to process. How to
		// distinguish long-running processes from those that have crashed?
		res := <-chRes
		if res.err != nil {
			fmt.Println(fmt.Errorf("extraction failed: %v", res.err))
			failed += 1
		} else {
			fmt.Println("Done:", res.wi.Outfile)
			successful += 1
		}
	}
	close(chJob) // causes workers to exit loop
	wg.Wait()    // wait workers
	return Status{Successful: successful, Failed: failed, Submitted: len(workItems)}
}

// Produce a printable string from Status
func (s Status) String() string {
	wtf := s.Submitted - (s.Successful + s.Failed)
	return fmt.Sprintf("total %d submitted jobs => success: %d, failed: %d, missing: %d", s.Submitted, s.Successful, s.Failed, wtf)
}
