package queue

import (
	"sync"
)

var (
	globalQueue *Queue
)

// Run start running queues,
// specify the number of buffers, and the number of worker threads
func Run(maxQueues, maxWorkers int) {
	if globalQueue == nil {
		globalQueue = NewQueue(maxQueues, maxWorkers)
	}
	globalQueue.Run()
}

// Push put the executable task into the queue
func Push(job Jober) {
	if globalQueue == nil {
		return
	}
	globalQueue.Push(job)
}

// Terminate terminate the queue to receive the task and release the resource
func Terminate() {
	if globalQueue == nil {
		return
	}
	globalQueue.Terminate()
}

// Queue a task queue for mitigating server pressure in high concurrency situations
// and improving task processing
type Queue struct {
	maxWorkers int
	jobQueue   chan Jober
	workerPool chan chan Jober
	workers    []Worker
	isRunning  bool
	wg         *sync.WaitGroup
	done       func()
}

// NewQueue create a queue that specifies the number of buffers and the number of worker threads
func NewQueue(maxQueues, maxWorkers int) *Queue {
	wg := new(sync.WaitGroup)
	return &Queue{
		jobQueue:   make(chan Jober, maxQueues),
		maxWorkers: maxWorkers,
		workerPool: make(chan chan Jober, maxWorkers),
		workers:    make([]Worker, maxWorkers),
		wg:         wg,
		done: func() {
			wg.Done()
		},
	}
}

// Run start running queues
func (q *Queue) Run() {
	if q.isRunning {
		return
	}

	q.isRunning = true
	for i := 0; i < q.maxWorkers; i++ {
		q.workers[i] = NewWorker(q.workerPool, q.done)
		q.workers[i].Start()
	}

	q.dispatcher()
}

func (q *Queue) dispatcher() {
	go func() {

		for job := range q.jobQueue {
			worker := <-q.workerPool
			worker <- job
		}

	}()
}

// Terminate terminate the queue to receive the task and release the resource
func (q *Queue) Terminate() {
	if !q.isRunning {
		return
	}

	q.isRunning = false

	q.wg.Wait()

	for i := 0; i < q.maxWorkers; i++ {
		q.workers[i].Terminate()
	}

	close(q.jobQueue)
	close(q.workerPool)
}

// Push put the executable task into the queue
func (q *Queue) Push(job Jober) {
	if !q.isRunning {
		return
	}

	q.wg.Add(1)
	q.jobQueue <- job
}
