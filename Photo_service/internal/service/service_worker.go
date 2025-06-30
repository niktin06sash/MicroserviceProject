package service

import (
	"log"
)

func (use *PhotoServiceImplement) taskWorker(i int) {
	defer use.wg.Done()
	for {
		select {
		case task, ok := <-use.Task_queue:
			if !ok {
				log.Printf("[INFO] [Photo-Service] [Worker: %v] Task channel closed, stopping worker", i)
				return
			}
			log.Printf("[INFO] [Photo-Service] [Worker: %v] New task has been received. Execute...", i)
			task()
		case <-use.closechan:
			return
		}
	}
}
func (use *PhotoServiceImplement) StopWorkers() {
	close(use.closechan)
	close(use.Task_queue)
	use.wg.Wait()
	log.Printf("[DEBUG] [Photo-Service] Successful stop task-workers")
}
