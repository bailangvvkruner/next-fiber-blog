package pool

import (
	"context"
	"sync"
	"sync/atomic"
)

type PanicHandler func(context.Context, any)

// Pool 协程池接口
type Pool interface {
	Name() string
	SetCap(cap int32)
	Go(fn func())
	CtxGo(ctx context.Context, fn func())
	DoGo(ctx context.Context, handler PanicHandler, fn func())
	WorkerCount() int32
}

type task struct {
	ctx     context.Context
	fn      func()
	handler PanicHandler
	next    *task
}

var taskPool sync.Pool

func init() {
	taskPool.New = func() any {
		return &task{}
	}
}

func (t *task) Recycle() {
	t.ctx = nil
	t.fn = nil
	t.next = nil
	taskPool.Put(t)
}

type simplePool struct {
	name        string
	cap         int32
	taskLock    sync.Mutex
	taskHead    *task
	taskTail    *task
	taskCount   int32
	workerCount int32
}

func NewPool(name string, cap int32) Pool {
	return &simplePool{
		name: name,
		cap:  cap,
	}
}

func (s *simplePool) Name() string {
	return s.name
}
func (s *simplePool) SetCap(cap int32) {
	atomic.StoreInt32(&s.cap, cap)
}
func (s *simplePool) Go(fn func()) {
	s.CtxGo(context.Background(), fn)
}
func (s *simplePool) CtxGo(ctx context.Context, fn func()) {
	s.DoGo(ctx, nil, fn)
}

func (s *simplePool) DoGo(ctx context.Context, handler PanicHandler, fn func()) {
	t := taskPool.Get().(*task)
	t.ctx = ctx
	t.handler = handler
	t.fn = fn
	s.taskLock.Lock()
	if s.taskHead == nil {
		s.taskHead = t
		s.taskTail = t
	} else {
		s.taskTail.next = t
		s.taskTail = t
	}
	s.taskLock.Unlock()
	atomic.AddInt32(&s.taskCount, 1)
	// 在待处理任务大于 0 并且以及启动的worker协程小于指定限制（cap）的情况下才会启动新的goroutine
	if atomic.LoadInt32(&s.taskCount) > 0 && atomic.LoadInt32(&s.workerCount) < atomic.LoadInt32(&s.cap) {
		s.addWorkCount()
		w := workerPool.Get().(*worker)
		w.pool = s
		w.run()
	}
}

func (s *simplePool) WorkerCount() int32 {
	return s.workerCount
}

// 添加worker线程
func (s *simplePool) addWorkCount() {
	atomic.AddInt32(&s.workerCount, 1)
}

// 删除工作线程
func (s *simplePool) subWorkCount() {
	atomic.AddInt32(&s.workerCount, -1)
}
