package threadpool

import (
	"sync"
)

type TaskQueue struct {
	pool  *ThreadPool
	queue map[string]*taskQueueStatusRunnable
	lock  sync.Mutex
}

func (q *TaskQueue) Submit(taskId string, runnable Runnable) *Future {
	q.lock.Lock()
	defer q.lock.Unlock()

	run := &taskQueueStatusRunnable{
		//callableTask: realTask,
		Runnable: runnable,
		taskId:   taskId,
		Status:   taskStatusQueue,
	}
	futureTask, future := q.pool.submitWithResult(run)
	run.futureTask = futureTask
	q.queue[taskId] = run

	go func() {
		<-futureTask.ctx.Done()
		q.lock.Lock()
		delete(q.queue, taskId)
		q.lock.Unlock()
	}()

	return future
}

func (q *TaskQueue) SetQueueCoreSize(core, bufferSize int) {
	q.lock.Lock()
	defer q.lock.Unlock()

	pool := q.pool
	pool.Stop()

	pool.bufferSize = bufferSize
	pool.coreSize = core
	pool.start()

	// 暂时想不到将旧队列移动到新队列的操作。
	// 想到了, 在任务的状态中设定一个标记，如果有这个标记的重新进入任务队列，那么直接在
	// worker 中阻塞后等待任务完成 return 即可
	for _, info := range q.queue {
		// taskQueueChan 的缓冲区很大，不用担心阻塞问题
		// 重启的时候将已提交的任务在提交一遍
		pool.channel <- info.futureTask
	}
}

// Len 活动的大小
func (q *TaskQueue) Len() int {
	return q.pool.WaitLen()
}

func NewTaskQueue(name string, coreSize int, bufferSize int) *TaskQueue {
	pool := NewThreadPool(name, coreSize, bufferSize)
	return &TaskQueue{
		pool:  pool,
		queue: make(map[string]*taskQueueStatusRunnable, coreSize),
		lock:  sync.Mutex{},
	}
}
