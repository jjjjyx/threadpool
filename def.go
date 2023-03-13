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
	future   *Future
}

// Future is the future returned after submitting a callable task to the thread ThreadPool
type Future struct {
	response <-chan interface{}
	done     bool
	err      error
	cancel   context.CancelFunc
	ctx      context.Context
}

func (f *Future) GetTaskContext() context.Context {
	return f.ctx
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
