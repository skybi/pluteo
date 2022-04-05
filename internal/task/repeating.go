package task

import "time"

// RepeatingTask executes a task in a specific interval asynchronously
type RepeatingTask struct {
	task     func()
	interval time.Duration

	running bool
	stop    chan struct{}
}

// NewRepeating creates a new repeating asynchronous task
func NewRepeating(task func(), interval time.Duration) *RepeatingTask {
	return &RepeatingTask{
		task:     task,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

// Start starts the repeating task.
// If the task is already running, this is a no-op.
func (task *RepeatingTask) Start() {
	if task.running {
		return
	}
	go func() {
		for {
			select {
			case <-time.After(task.interval):
				task.task()
			case <-task.stop:
				return
			}
		}
	}()
	task.running = true
}

// Stop stops the repeating task.
// If the task is not running, this is a no-op.
func (task *RepeatingTask) Stop() {
	if !task.running {
		return
	}
	close(task.stop)
	task.running = false
}
