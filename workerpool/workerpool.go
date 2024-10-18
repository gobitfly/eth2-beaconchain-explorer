package workerpool

import (
	"sync"
)

type WorkerPool struct {
	maxWorker   int
	queuedTaskC chan func()
	quitChan    chan bool
	waitGroup   sync.WaitGroup
}

// New will create an instance of WorkerPool.
func New(workers, size int) *WorkerPool {
	return &WorkerPool{
		maxWorker:   workers,
		queuedTaskC: make(chan func(), size),
		quitChan:    make(chan bool),
		waitGroup:   sync.WaitGroup{},
	}
}

// AddTask adds a function to the queued task channel
func (wp *WorkerPool) AddTask(task func()) {
	wp.waitGroup.Add(1)
	wp.queuedTaskC <- task
}

// Run starts the WorkerPool. Tasks added to the queued task channel will be executed
// until the quit channel returns true
func (wp *WorkerPool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		go worker(&wp.waitGroup, wp.queuedTaskC, wp.quitChan)
	}
}

// TotalQueuedTask returns the total tasks left in the queue
func (wp *WorkerPool) TotalQueuedTask() int {
	return len(wp.queuedTaskC)
}

// Quit will close all worker go routines
func (wp *WorkerPool) Quit() {
	for i := 0; i < wp.maxWorker; i++ {
		wp.quitChan <- true
	}
	close(wp.queuedTaskC)
}

func (wp *WorkerPool) Wait() {
	wp.waitGroup.Wait()
}

func worker(wg *sync.WaitGroup, jobs <-chan func(), quit chan bool) {
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			job()
			wg.Done()
		case <-quit:
			return
		}
	}
}
