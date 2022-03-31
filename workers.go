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

func Process(workItems []workItem, maxConcurrent int) {
	if maxConcurrent <= 0 {
		maxConcurrent = runtime.NumCPU()
	}

	var wg sync.WaitGroup
	chJob := make(chan job)
	chRes := make(chan result, len(workItems))
	go func() {
		for res := range chRes {
			if res.err != nil {
				fmt.Println(fmt.Errorf("extraction failed: %v", res.err))
			} else {
				fmt.Println("Done:", res.wi.Outfile)
			}
		}
	}()
	for t := 0; t < maxConcurrent; t++ {
		wg.Add(1)
		go worker(chJob, chRes, &wg)
	}
	for i := range workItems {
		//fmt.Printf("%+v\n", wi.FFmpegArgs())
		chJob <- job{&workItems[i]}
	}
	close(chJob)
	wg.Wait()
	close(chRes)
}

func worker(jobs <-chan job, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		err := job.wi.Process()
		results <- result{job.wi, err}
	}
}
