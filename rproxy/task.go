package rproxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/hashicorp/go-uuid"
)

type callResult struct {
	ret interface{}
	err error
}

type CallTask struct {
	isRunning int32
	id        string
	startAt   time.Time

	// 底层调用参数相关
	ctx        context.Context
	paramsLock sync.Mutex
	params     []interface{}
	methodName string
	result     callResult

	wg         *sync.WaitGroup
	windowSize int
	timeoutAt  time.Time
}

func (c *CallTask) String() string {
	return fmt.Sprintf("CallTask(id='%s', name='%s')", c.id, c.methodName)
}

func NewCallTask(ctx context.Context, methodName string, windowSize int, timeout time.Duration) *CallTask {
	id, _ := uuid.GenerateUUID()
	return &CallTask{
		id:         id,
		startAt:    time.Now(),
		ctx:        ctx,
		params:     make([]interface{}, 0),
		wg:         &sync.WaitGroup{},
		methodName: methodName,
		windowSize: windowSize,
		timeoutAt:  time.Now().Add(timeout),
	}
}

func (c *CallTask) ShouldExecuteNow() bool {
	c.paramsLock.Lock()
	defer c.paramsLock.Unlock()

	return len(c.params) >= c.windowSize || time.Now().After(c.timeoutAt)
}

func (c *CallTask) EnqueueAndWait(param interface{}, callback ExtractResultFunc) (interface{}, error) {
	log.Infof("[%s][EnqueueAndWait] enqueue param %v for method '%s'", c, param, c.methodName)
	c.paramsLock.Lock()
	c.params = append(c.params, param)
	c.paramsLock.Unlock()

	// 提供参数的 goroutine 安心睡觉吧
	c.wg.Add(1)
	c.wg.Wait()

	log.Infof("[%s][EnqueueAndWait] wakeup now, get result from method '%s' for param %v", c, c.methodName, param)
	// 唤醒后，记得提取需要的结果
	if c.result.err != nil {
		return nil, c.result.err
	}

	return callback(c.result.ret, param), nil
}

func (c *CallTask) Execute() {
	fn, ok := rpcMapping[c.methodName]
	if !ok {
		panic(fmt.Sprintf("method '%s' not registered", c.methodName))
	}

	log.Infof("[%s][Execute] call method '%s'", c, c.methodName)
	defer func() {
		// 一定要唤醒等待者
		log.Infof("[%s][Execute] call method '%s' finished, wakeup waiters", c, c.methodName)
		c.wg.Add(-len(c.params))
	}()

	// TODO: 以后可以加上通用降级逻辑
	r, err := fn(c.ctx, c.params...)
	c.result = callResult{ret: r, err: err}
}
