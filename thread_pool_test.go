package threadpool

import (
	"context"
	"errors"
	"fmt"
	"time"

	"testing"
)

type TestErrorTask struct {
	err error
}

func (t TestErrorTask) Run(context.Context) {
	//panic("implement me")
	panic(t.err)
}

type TestSleepTask struct {
	t int
}

func (t *TestSleepTask) Run(context.Context) {
	//fmt.Println("start")
	time.Sleep(time.Duration(t.t) * time.Second)
	//fmt.Println("end")
}

type TestTask struct {
	TestData *TestData
}

type TestData struct {
	Val string
}

func (t *TestTask) Run(context.Context) {
	fmt.Println("Running the task")
	t.TestData.Val = "changed"
}

type TestTaskFuture struct {
	t int
}

func (t *TestTaskFuture) Call(ctx context.Context) interface{} {

	// ctx cancel
	// 这里示范下 ctx 取消的时候 如何中断当前执行

	timer := time.NewTimer(time.Duration(t.t) * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return "cccc"
	case <-timer.C:
		return "end"
	}

}

func TestNewThreadPool(t *testing.T) {
	threadPool := NewThreadPool("test", 3, 3*100)
	fmt.Println(threadPool)
}

func TestNewThreadPool_Execute(t *testing.T) {

	data := &TestData{Val: "pristine"}
	task := &TestTask{TestData: data}

	pool := NewThreadPool("test", 3, 3*100)
	pool.Submit(task)

	time.Sleep(2 * time.Second)
	fmt.Println("")
	fmt.Println("data.Val", data.Val)
	if data.Val != "changed" {
		t.Fail()
	}
}

func TestThreadPool_SubmitSuccess(t *testing.T) {
	task := &TestTaskFuture{}

	pool := NewThreadPool("test", 3, 3*100)
	future := pool.SubmitWithResult(task)

	time.Sleep(3 * time.Second)

	response, err := future.Get()

	if !future.Done() {
		t.Fatalf("future status should done")
	}

	if err != nil {
		fmt.Println(err)
		t.Fatalf("future err should nil")
	}

	if response != "end" {
		t.Fatalf("future ret should equals 'end'")
	}

	fmt.Println("Thread done resp =", response)
}

func TestThreadPool_SubmitCancel(t *testing.T) {
	task := &TestTaskFuture{
		t: 3,
	}

	pool := NewThreadPool("test", 3, 3*100)
	future := pool.SubmitWithResult(task)

	go func() {
		time.Sleep(time.Second)
		future.Cancel()
	}()

	response, err := future.Get()

	if !future.Done() {
		t.Fatalf("future status should done")
	}

	if err == nil {
		t.Fatalf("future err should not nil")
	}

	if response != nil {
		t.Fatalf("future ret should equels nil")
	}

	fmt.Println("Thread done err =", err)
}

func TestThreadPool_panic(t *testing.T) {
	c := 3
	b := 300

	pool := NewThreadPool("test_cap", c, b)

	t2 := &TestErrorTask{err: errors.New("假设这里有一个错误")}
	future := pool.Submit(t2)
	time.Sleep(time.Second)

	if !future.Done() {
		t.Fatalf("future status should done")
	}
	response, err := future.Get()

	if response != nil {
		t.Fatalf("future ret should equels nil")
	}

	if err != t2.err {
		t.Fatalf("future err should euqals err")
	}

}

func TestThreadPool_Cap(t *testing.T) {
	c := 1
	b := 300

	pool := NewThreadPool("test_cap", c, b)

	pool.Submit(&TestSleepTask{1})
	pool.Submit(&TestSleepTask{20})
	pool.Submit(&TestSleepTask{30})
	pool.Submit(&TestSleepTask{40})

	//time.Sleep(1500)
	//go func() {
	//	for {
	//		fmt.Println("xxxxxThread done WaitLen =", pool.WaitLen())
	//		time.Sleep(time.Second)
	//	}
	//}()

	if pool.Cap() != b {
		t.Fatalf("Pool Cap() = %d; expected %d", pool.Cap(), b)
	}

	if pool.Core() != c {
		t.Fatalf("Pool Core() = %d; expected %d", pool.Core(), c)
	}

	//if pool.WaitLen() != 3 {
	//	t.Fatalf("Pool WaitLen() = %d; expected %d", pool.WaitLen(), 3)
	//}

	fmt.Println("Thread done Cap =", pool.Cap())
	fmt.Println("Thread done WaitLen =", pool.WaitLen())
	fmt.Println("Thread done Core =", pool.Core())

}
