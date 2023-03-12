package threadpool

import (
	"context"
	"errors"
	"time"
)

// Callable
// 做程序设计的时候又想起泛型的好了
// 嗨。。。
type Callable interface {
	Call(context.Context) interface{}
}

type Runnable interface {
	Run(context.Context)
}

// 转换 Runnable 成 Callable
type notResultCallable struct {
	Runnable
}

func (t *notResultCallable) Call(ctx context.Context) interface{} {
	t.Run(ctx)
	return nil
}

// callableTask is internally used to wrap the callable and future together
// So that the worker can send the response back through channel provided in Future object
type callableTask struct {
	Callable
	ctx      context.Context
	cancel   context.CancelFunc
	response chan<- interface{}
	handle   *Future
	before   func() bool
}

//type taskQueueCallable

//type taskStatus int
const (
	taskStatusQueue   = iota // 队列中
	taskStatusRunning        // 运行中
	taskStatusEnd            // 结束

)

// 带有任务状态的任务执行类
// 本来想结构体名字带有 abstract 想想不够合理，这里的虽然是实现了  Runnable 接口进行扩展，但是其中 Runnable 会由具体的实现进行设置
// 顶对顶多算是一个装饰器
type taskQueueStatusRunnable struct {
	Runnable
	taskId     string
	Status     int
	futureTask callableTask
}

func (task *taskQueueStatusRunnable) Call(ctx context.Context) interface{} {
	if task.running() {
		//fmt.Println(task.taskId, "runningxxxxxx")
		// 这里缺少一个打印标记。。 只能不打印了
		// !!! 这里也不太应该用日志打印..!!
		// logrus.Tracef("任务[%s]已在队列中执行，阻塞当前队列等到任务完成后退出", task.taskId)
		<-ctx.Done()
		return nil
	}

	if task.end() {
		// logrus.Tracef("任务[%s]退出执行队列, 任务状态=%d,无需重复执行。", task.taskId, task.Status)
		return nil
	}

	task.Status = taskStatusRunning
	task.Run(ctx)
	task.Status = taskStatusEnd
	//fmt.Println(task.taskId, "Call end")
	return nil

}

func (task *taskQueueStatusRunnable) running() bool {
	return task.Status == taskStatusRunning
}

func (task *taskQueueStatusRunnable) end() bool {
	return task.Status == taskStatusEnd
}

// Future is the handle returned after submitting a callable task to the thread ThreadPool
type Future struct {
	response <-chan interface{}
	done     bool
	err      error
	cancel   context.CancelFunc
}

// Get returns the response of the Callable task when done
// the blocking call it waits for the execution to complete

func (f *Future) Get() (interface{}, error) {

	if f.err != nil {
		return nil, f.err
	}

	ret, ok := <-f.response
	if !ok {
		return nil, context.Canceled
	}

	return ret, nil
}

func (f *Future) GetWithTimeout(timeout time.Duration) (interface{}, error) {

	if f.err != nil {
		return nil, f.err
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case ret, ok := <-f.response:
		if !ok {
			return nil, context.Canceled
		}
		return ret, nil
	case <-timer.C:
		return nil, errors.New("timeout")
	}
}

// Done returns true if the execution is already done
func (f *Future) Done() bool {
	return f.done
}

func (f *Future) Cancel() bool {

	if f.cancel != nil {
		f.cancel()
		f.err = context.Canceled
		return true
	}
	return false
}
