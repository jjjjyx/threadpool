package example

import (
	"context"
	"fmt"
	"threadpool"
	"time"
)

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

func method1() {
	data := &TestData{Val: "pristine"}
	task := &TestTask{TestData: data}

	pool := threadpool.NewThreadPool("test", 3, 3*100)
	pool.Submit(task)

	time.Sleep(2 * time.Second)
	fmt.Println("")
	fmt.Println("data.Val", data.Val)
}
