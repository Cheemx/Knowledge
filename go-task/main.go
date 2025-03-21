/*

Problem Statement:
Design a concurrent task processor in Go that:

Spawns a configurable number of worker goroutines (N).

Accepts tasks (jobs) via a channel.

Processes jobs concurrently (simulate work with a sleep).

Shuts down gracefully when all jobs are processed.

Requirements:

Use goroutines, channels, and synchronization primitives (e.g., sync.WaitGroup).

Avoid goroutine leaks.

Ensure all workers exit cleanly after processing all jobs.


*/

package main

import (
	"time"
)

type Job struct {
	ID       int
	Duration time.Duration // Simulated processing time
}

func main() {
	const numWorkers = 3
	const numJobs = 5

	// TODO: Create a buffered channel for jobs.
	// TODO: Start N worker goroutines.
	// TODO: Enqueue jobs into the channel (e.g., IDs 1-5 with increasing durations).
	// TODO: Gracefully close the channel and wait for workers to finish.
}

// TODO: Implement the worker function.
// Workers should:
// - Receive jobs from the channel.
// - Simulate work with a sleep (use job.Duration).
// - Print logs (e.g., "Worker X processing job Y").
