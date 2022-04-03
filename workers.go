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
	chJob := make(chan job)
	chRes := make(chan result)
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
