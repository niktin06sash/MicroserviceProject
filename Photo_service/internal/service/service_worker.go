package service

import (
	"context"
	"log"
	"sync"
)

type Worker struct {
	task_queue chan func()
	ctx        context.Context
	wg         *sync.WaitGroup
}

func NewWorker(ctx context.Context) *Worker {
	return &Worker{
		task_queue: make(chan func(), 1000),
		ctx:        ctx,
		wg:         &sync.WaitGroup{},
	}
}
func (wp *Worker) Start() {
	for i := 1; i <= 5; i++ {
		wp.wg.Add(1)
		go wp.taskWorker(i)
	}
}
func (wp *Worker) taskWorker(i int) {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.task_queue:
			if !ok {
				log.Printf("[INFO] [Photo-Service] [Worker: %v] Task channel closed, stopping Task-worker", i)
				return
			}
			log.Printf("[INFO] [Photo-Service] [Worker: %v] New task has been received. Execute...", i)
			task()
		case <-wp.ctx.Done():
			log.Printf("[DEBUG] [Photo-Service] [Worker: %v] Context canceled, stopping Task-worker...", i)
			return
		}
	}
}
func (wp *Worker) Stop() {
	close(wp.task_queue)
	wp.wg.Wait()
	log.Printf("[DEBUG] [Photo-Service] Successful stop Task-workers")
}
