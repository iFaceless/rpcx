package rproxy

import (
	"context"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	windowSize    int
	windowTimeout time.Duration

	// 正在等待执行的批次
	tasks sync.Map

	stop chan struct{}
}

type Option func(proxy *Proxy)

func WindowSize(n int) Option {
	return func(proxy *Proxy) {
		proxy.windowSize = n
	}
}

func WindowTimeout(t time.Duration) Option {
	return func(proxy *Proxy) {
		if t < 1*time.Millisecond {
			panic("minimum timeout is 1ms")
		}

		proxy.windowTimeout = t
	}
}

func (p *Proxy) String() string {
	return "Proxy"
}

func NewProxy(opts ...Option) *Proxy {
	p := &Proxy{
		windowSize:    10,
		windowTimeout: time.Millisecond * 1, // 1ms
		stop:          make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Proxy) Serve() {
	log.Infof("Proxy started...")

	listTasks := func() []*CallTask {
		tasks := make([]*CallTask, 0)
		p.tasks.Range(func(key, value interface{}) bool {
			tasks = append(tasks, value.(*CallTask))
			return true
		})
		return tasks
	}

	execTask := func(task *CallTask) {
		task.Execute()
	}

	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infof("[%s] I'm alive", p)
		case <-p.stop:
			log.Infof("[%s] Shutdown proxy...", p)
			return
		default:
			tasks := listTasks()
			if len(tasks) == 0 {
				//time.Sleep(10 * time.Microsecond) // idle
				break
			}

			for _, t := range tasks {
				if t.ShouldExecuteNow() {
					p.tasks.Delete(t.methodName)
					go execTask(t)
				}
			}
		}
	}
}

func (p *Proxy) Shutdown() {
	p.stop <- struct{}{}
	close(p.stop)
}

func (p *Proxy) Call(ctx context.Context, methodName string, param interface{}, extractFn ExtractResultFunc) (interface{}, error) {
	// 简单起见，假设批量调用的函数就是在原来函数名的基础上加个 Batch 前缀
	if !strings.HasPrefix(methodName, "Batch") {
		methodName = "Batch" + methodName
	}

	val, _ := p.tasks.LoadOrStore(methodName, NewCallTask(ctx, methodName, p.windowSize, p.windowTimeout))
	task := val.(*CallTask)
	return task.EnqueueAndWait(param, extractFn)
}
