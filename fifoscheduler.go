package scheduler

import (
	"context"
	"errors"
	"sync"
)

// fifo represent as FIFO scheduler.
type fifo struct {
	mu sync.Mutex
	resume    chan struct{}
	scheduled int
	finished  int
	pendings  []Job
	ctx    context.Context
	cancel context.CancelFunc
	finishCond *sync.Cond
	donec      chan struct{}
}

// NewFIFOScheduler returns a Scheduler that schedules jobs in FIFO
// order sequentially
func NewFIFOScheduler() Scheduler {
	f := &fifo{
		resume: make(chan struct{}, 1),
		donec:  make(chan struct{}, 1),
	}
	f.finishCond = sync.NewCond(&f.mu)
	f.ctx, f.cancel = context.WithCancel(context.Background())
	go f.run()
	return f
}

// Schedule schedules a job that will be ran in FIFO order sequentially.
func (f *fifo) Schedule(j Job) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cancel == nil {
		return errors.New("schedule to stopped scheduler")
	}

	if len(f.pendings) == 0 {
		select {
		case f.resume <- struct{}{}:
		default:
		}
	}
	f.pendings = append(f.pendings, j)
	return nil
}

func (f *fifo) Pending() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.pendings)
}

func (f *fifo) Scheduled() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.scheduled
}

func (f *fifo) Finished() int {
	f.finishCond.L.Lock()
	defer f.finishCond.L.Unlock()
	return f.finished
}

func (f *fifo) WaitFinish(n int) {
	f.finishCond.L.Lock()
	for f.finished < n || len(f.pendings) != 0 {
		f.finishCond.Wait()
	}
	f.finishCond.L.Unlock()
}

// Stop stops the scheduler and cancels all pending jobs.
func (f *fifo) Stop() {
	f.mu.Lock()
	f.cancel()
	f.cancel = nil
	f.mu.Unlock()
	<-f.donec
}

func (f *fifo) run() {
	// TODO: recover from job panic?
	defer func() {
		close(f.donec)
		close(f.resume)
	}()

	for {
		var task Job
		f.mu.Lock()
		if len(f.pendings) != 0 {
			f.scheduled++
			task = f.pendings[0]
		}
		f.mu.Unlock()
		if task == nil {
			select {
			case <-f.resume:
			case <-f.ctx.Done():
				// naive way to "handle" stop pending tasks.
				f.mu.Lock()
				pendings := f.pendings
				f.pendings = nil
				f.mu.Unlock()
				// clean up pending jobs
				for _, todo := range pendings {
					todo(f.ctx)
				}
				return
			}
		} else {
			task(f.ctx)
			f.finishCond.L.Lock()
			f.finished++
			f.pendings = f.pendings[1:]
			f.finishCond.Broadcast()
			f.finishCond.L.Unlock()
		}
	}
}