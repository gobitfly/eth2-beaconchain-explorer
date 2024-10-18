package workerpool

import (
	"runtime"
	"testing"
	"time"
)

func Test_WorkerPool(t *testing.T) {
	startThreads := runtime.NumGoroutine()

	totalWorker := 5
	sleepTime := 5 * time.Millisecond
	totalTask := 1000
	wp := New(totalWorker, totalTask)

	completedTaskChannel := make(chan int, totalTask)
	defer close(completedTaskChannel)

	for i := 0; i < totalTask; i++ {
		taskID := i + 1
		wp.AddTask(func() {
			completedTaskChannel <- taskID
			time.Sleep(sleepTime)
		})
	}
	start := time.Now()

	wp.Run()
	wp.Wait()
	if got, want := len(completedTaskChannel), totalTask; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	elapsed := time.Since(start)

	if got, want := runtime.NumGoroutine(), startThreads+totalWorker; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	if got, want := float64(elapsed.Milliseconds()), float64(sleepTime.Milliseconds()*int64(totalTask)/int64(totalWorker))*1.4; got > want {
		t.Errorf("unexpected execution time, expected t < %vms, got t=%vms", want, got)
	}

	wp.Quit()
	time.Sleep(1 * time.Millisecond)
	endThreads := runtime.NumGoroutine()
	if startThreads != endThreads {
		t.Errorf("unexpected go thread count: got %v, want %v", endThreads, startThreads)
	}
}
