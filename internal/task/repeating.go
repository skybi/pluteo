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
	}
}

// Start starts the repeating task.
// If the task is already running, this is a no-op.
// A call to Stop as soon as the task is no longer needed is highly recommended as the object would not be garbage
// collected otherwise.
func (task *RepeatingTask) Start() {
	if task.running {
		return
	}
	task.stop = make(chan struct{})
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
// forceExec defines whether to execute the task one last time just before the task shuts down.
func (task *RepeatingTask) Stop(forceExec bool) {
	if !task.running {
		return
	}
	close(task.stop)
	task.running = false
	if forceExec {
		task.task()
	}
}
