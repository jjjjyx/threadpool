# Golang Thread Pool
[![GoDoc](https://godoc.org/github.com/jjjjyx/threadpool?status.svg)](https://godoc.org/github.com/jjjjyx/threadpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjjjyx/threadpool)](https://goreportcard.com/report/github.com/jjjjyx/threadpool)

[![My Skills](https://skillicons.dev/icons?i=go,github,idea)](https://skillicons.dev)

主要特性：

* 支持 `Callable` 和 `Runnable`，支持 `Future` 或任务执行结果
* `Future` 获取支持超时时间
* 支持对任务过程错误的捕获，并能在 `Future` 获取错误
* 仿照`Java` 线程池设计


## Thread Pool 使用

### 为需要执行的任务实现 `Runnable` or  `Callable` 接口。 例如

```go

type MyTask struct { }

func (t *MyTask) Run(ctx context.Context){
// Do your task here
}


type MyTask2 struct { }

func (t *MyTask2) Call() interface{} {
    // Do your task here
}
```

### 创建线程对象

```go

pool := threadpool.NewThreadPool("defaultTaskPool", 3, 3*100)

```

### 提交任务
```go
future1 := pool.Submit(&MyTask{})

future2 := pool.SubmitWithResult(&MyTask2{})

// 任务是否完成
done := future2.Done

// 获取任务任务结果
ret, err := future2.Get()

// 获取任务任务结果指定超时时间
ret, err := future2.GetWithTimeout(3 * time.Second)

// 取消任务
future2.Cancel()
future1.Cancel()

```

### stop

```go
// 停止任务拒绝掉所有没有完成的任务
pool.Stop()

// 埋个坑 有人看在实现 awaitTermination
```
