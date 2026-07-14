// Command 08_worker_pool shows a bounded worker pool: a fixed number of
// worker goroutines pull jobs from a shared channel, so the amount of
// concurrent work is capped regardless of how many jobs there are.
package main

import (
	"fmt"
	"sort"
	"sync"
)

// job is one unit of work: square the number N.
type job struct {
	id int
	n  int
}

// result is what a worker produces for a given job.
type result struct {
	id    int
	value int
}

// worker reads jobs from jobs until it is closed, computes each result, and
// sends it to results. It never closes results itself, because it is one of
// possibly several workers sharing that channel; only the coordinator that
// knows when every worker has stopped may close it safely.
func worker(jobs <-chan job, results chan<- result) {
	for j := range jobs {
		results <- result{id: j.id, value: j.n * j.n}
	}
}

// runWorkerPool computes n*n for every input using a bounded pool of
// workerCount goroutines, returning results sorted by job id for
// deterministic output. workerCount lets you bound concurrency (and
// therefore resource usage) independently of how many jobs there are.
func runWorkerPool(inputs []int, workerCount int) []result {
	jobs := make(chan job, len(inputs))
	results := make(chan result, len(inputs))

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			worker(jobs, results)
		}()
	}

	for i, n := range inputs {
		jobs <- job{id: i, n: n}
	}
	close(jobs) // no more jobs; lets every worker's range loop end

	// Close results only after every worker has stopped sending to it.
	// Doing the wait in its own goroutine lets the range below start
	// draining results immediately instead of waiting for all workers to
	// finish before receiving anything.
	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]result, 0, len(inputs))
	for r := range results {
		collected = append(collected, r)
	}

	sort.Slice(collected, func(i, j int) bool { return collected[i].id < collected[j].id })
	return collected
}

func main() {
	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	results := runWorkerPool(inputs, 3) // never more than 3 jobs in flight
	for _, r := range results {
		fmt.Printf("job %d -> %d\n", r.id, r.value)
	}
}
