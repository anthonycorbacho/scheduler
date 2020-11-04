package scheduler

import "context"

// Job represent a func that will be executed by a Scheduler
type Job func(context.Context)

// Scheduler can schedule jobs.
type Scheduler interface {
	// Schedule asks the scheduler to schedule a job defined by the given func.
	Schedule(j Job) error

	// Pending returns number of pending jobs
	Pending() int

	// Scheduled returns the number of scheduled jobs (excluding pending jobs)
	Scheduled() int

	// Finished returns the number of finished jobs
	Finished() int

	// WaitFinish waits until at least n job are finished and all pending jobs are finished.
	WaitFinish(n int)

	// Stop stops the scheduler.
	Stop()
}