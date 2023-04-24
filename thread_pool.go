package threadpool

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
)

type ThreadPool struct {
	coreSize   int
	bufferSize int
	channel    chan callableTask
	name       string
}

func (t *ThreadPool) start() {
	if t.channelIsOpen() {
		return
	}
	t.channel = make(chan callableTask, t.bufferSize)
	for i := 0; i < t.coreSize; i++ {
		go func(i int) {
			t.worker(i)
		}(i)
	}
}

// Shutdown 与close 的区别在于 停止后是否要中断channel缓冲区的任务
func (t *ThreadPool) Shutdown() {
	if !t.channelIsOpen() {
		return
	}

	close(t.channel)
	t.channel = nil
}

// Stop 停止线程池，不在接受任务，已开始的任务不处理，已在缓冲区的任务取消
func (t *ThreadPool) Stop() {
	if !t.channelIsOpen() {
		return
	}

	close(t.channel)
	for _ = range t.channel {
		// 空接受, 清空掉旧的
	}
	t.channel = nil
}

func (t *ThreadPool) channelIsOpen() bool {
	if t.channel == nil {
		return false
	}

	select {
	case _, ok := <-t.channel:
		if !ok {
			return false
		}
		return true
	default:
		return true
	}
}

func (t *ThreadPool) worker(threadId int) {
	// ::思考::
	// 如果想做任务取消的话，需要有个ctx 上下文
	// 这个上下文对象 放在 Runnable 或 Callable 不太行
	// 应它两是接口
	// 如果继续包装 只能是将 Runnable 或 Callable 包装成一个 DefaultAbstractRunnable 这种东西
	// 就像现在的 callableTask 的实现

	worker := func(info callableTask) {
		defer func() {
			if re := recover(); re != nil {
				message := fmt.Sprintf("pool[%s][thread-%d]\nstacktrace from panic: %s: %s\n", t.name, threadId, re, string(debug.Stack()))
				_, _ = fmt.Fprintln(os.Stderr, message)
				if err, ok := re.(error); ok {
					info.future.err = err
				}
				info.future.done = true
			}

			info.cancel()
			close(info.response)
		}()

		ctx := info.ctx
		result := info.Call(ctx)
		info.future.done = true

		select {
		case <-ctx.Done():
			// 先被结束了，close 掉 不返回结果
		default:
			info.response <- result
		}

	}

	for info := range t.channel {
		worker(info)
	}

}

func (t *ThreadPool) submitTask(task callableTask) {
	t.channel <- task
}

func (t *ThreadPool) Submit(task Runnable) *Future {
	return t.SubmitWithResult(&notResultCallable{
		Runnable: task,
	})
}

func (t *ThreadPool) submitWithResult(task Callable) (callableTask, *Future) {
	ctx, cancel := context.WithCancel(context.Background())

	// 需要一个缓冲区，防止 worker 阻塞
	response := make(chan interface{}, 1)

	future := &Future{response: response, cancel: cancel, ctx: ctx}
	futureTask := callableTask{
		Callable: task,
		future:   future,
		ctx:      ctx,
		cancel:   cancel,
		response: response,
	}

	t.submitTask(futureTask)
	return futureTask, future
}

func (t *ThreadPool) SubmitWithResult(task Callable) *Future {
	_, future := t.submitWithResult(task)
	return future
}

// Cap 缓冲区大小
func (t *ThreadPool) Cap() int {
	return cap(t.channel)
}

// WaitLen 等待中的大小
func (t *ThreadPool) WaitLen() int {
	// 最新的感悟，len 中只存放不被 接受的部分
	// 所以len 一定是 队列大小
	return len(t.channel)
}

// Core 核心线程大小
// 排队的格式应当是 cap - core 大小 , 单独算吧 不提供实现了
func (t *ThreadPool) Core() int {
	return t.coreSize
}

func NewThreadPool(name string, coreSize int, bufferSize int) *ThreadPool {

	// 任务队列的缓存数据在想要不要保存在这里？
	// 设计上应该不应该，
	// 这里也只是一个 线程池罢了，如果带着线程队列， 那就耦合了
	// todo 考虑加上一个线程执行的超时时间

	threadPool := &ThreadPool{
		name:       name,
		coreSize:   coreSize,
		bufferSize: bufferSize,
		//channel:    taskQueueChan,
	}
	threadPool.start()

	return threadPool
}
