# Golang Thread Pool

[![My Skills](https://skillicons.dev/icons?i=go,github,idea)](https://skillicons.dev)

主要特性：

* 支持 `Callable` 和 `Runnable`，支持 `Future` 或任务执行结果
* `Future` 获取支持超时时间
* 支持对任务过程错误的捕获，并能在 `Future` 获取错误
* 仿照`Java` 线程池设计
* 任务队列支持查询任务队列大小，支持动态修改核心线程数

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

## Task Queue 任务执行队列
这个本质就是线程池，只是加入了一个 允许重置核心队列大小功能。

### 创建队列
```go
queue := NewTaskQueue("test", 10, 333)
```


### 提交任务 执行
```go
_ = queue.Submit("任务1", &MyTask{})
```

### 修改核心队列大小
修改核心线程大小，保持当前正在执行的任务队列，其主要实现目的：
比如原来队列 大小5个，修改成10，如果此时添加10 任务 会让总线程数超过 10
 这里的修改解决了这个问题，如果扩大线程池大小，会将旧队列中已执行的任务数量锁住新的队列数量，
如果修改的更小，不会停止正在运行的任务。

```go
queue.SetQueueCoreSize(3, 111)

```

