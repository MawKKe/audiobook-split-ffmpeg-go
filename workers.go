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
type Status struct {
	Successful int
	Failed     int
	Submitted  int
}

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

func (s Status) String() string {
	wtf := s.Submitted - (s.Successful + s.Failed)
	return fmt.Sprintf("total %d submitted jobs => success: %d, failed: %d, missing: %d", s.Submitted, s.Successful, s.Failed, wtf)
}
