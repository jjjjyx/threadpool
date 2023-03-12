package threadpool

import (
	"fmt"
	"testing"
	"time"
)

func TestTaskQueue_Submit(t *testing.T) {

	queue := NewTaskQueue("test", 1, 10)
	future := queue.Submit("任务一", &TestSleepTask{3})
	if queue.Len() != 1 {
		t.Fatalf("1. queue WaitLen() = %d; expected %d", queue.Len(), 1)
	}

	response, err := future.Get()
	if !future.Done() {
		t.Fatalf("future status should done")
	}

	if err != nil {
		t.Fatalf("future err should not nil")
	}

	if response != nil {
		t.Fatalf("future ret should equels nil")
	}

	fmt.Println("Thread done err =", err)
}

func TestTaskQueue_SetQueueCoreSize(t *testing.T) {
	queue := NewTaskQueue("test", 1, 333)

	// 奇怪啊， WaitLen 就是少一个！
	_ = queue.Submit("任务1", &TestSleepTask{3})
	_ = queue.Submit("任务2", &TestSleepTask{3})
	_ = queue.Submit("任务3", &TestSleepTask{3})
	_ = queue.Submit("任务4", &TestSleepTask{3})
	_ = queue.Submit("任务5", &TestSleepTask{3})
	_ = queue.Submit("任务6", &TestSleepTask{3})

	go func() {
		for {
			fmt.Println("Thread done WaitLen =", queue.Len(), queue.pool.Core())
			time.Sleep(time.Second)
		}
	}()

	//if queue.pool.Cap() != 333 {
	//	t.Fatalf("Pool Cap() = %d; expected %d", queue.pool.Cap(), 333)
	//}
	//
	//if queue.pool.Core() != 1 {
	//	t.Fatalf("Pool Core() = %d; expected %d", queue.pool.Core(), 1)
	//}
	//
	//if queue.WaitLen() != 4 {
	//	t.Fatalf("1. queue WaitLen() = %d; expected %d", queue.WaitLen(), 4)
	//}
	//
	//fmt.Println("Thread done Cap =", queue.pool.Cap())
	//fmt.Println("Thread done WaitLen =", queue.pool.WaitLen())
	//fmt.Println("Thread done Core =", queue.pool.Core())
	//
	//fmt.Println(queue.pool.Cap())
	time.Sleep(2 * time.Second)

	queue.SetQueueCoreSize(3, 111)

	time.Sleep(1000 * time.Second)

	//
	//time.Sleep(2000 * time.Millisecond)
	//fmt.Println(queue.pool.Cap())
	//if queue.WaitLen() != 1 {
	//	t.Fatalf("2 queue WaitLen() = %d; expected %d", queue.WaitLen(), 1)
	//}
}
