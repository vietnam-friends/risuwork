package workerx

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/isucon/isucandar/worker"
)

type Worker struct {
	mu                 sync.RWMutex
	w                  *worker.Worker
	workFunc           worker.WorkerFunc
	parallelism        int32
	parallelismUpdater func(int) int32
	opts               []worker.WorkerOption
}

func MustNewWorker(f worker.WorkerFunc, opts ...WorkerOption) *Worker {
	w, err := NewWorker(f, opts...)
	if err != nil {
		panic(err)
	}
	return w
}

func NewWorker(f worker.WorkerFunc, opts ...WorkerOption) (*Worker, error) {
	wx := &Worker{
		mu:                 sync.RWMutex{},
		workFunc:           f,
		parallelism:        -1,
		parallelismUpdater: func(l int) int32 { return int32(l) },
		opts:               make([]worker.WorkerOption, 0),
	}

	for _, opt := range opts {
		err := opt(wx)
		if err != nil {
			return nil, err
		}
	}

	w, err := worker.NewWorker(wx.workFunc, wx.opts...)
	if err != nil {
		return nil, err
	}
	wx.w = w
	return wx, nil
}

func (w *Worker) UpdateParallelism(newLv int) (int32, int32) {
	w.mu.Lock()
	defer w.mu.Unlock()

	old := w.parallelism
	w.parallelism = w.parallelismUpdater(newLv)
	w.w.SetParallelism(w.parallelism)
	return old, w.parallelism
}

func (w *Worker) Process(ctx context.Context) {
	w.w.Process(ctx)
}

func (w *Worker) Wait() {
	w.w.Wait()
}

func (w *Worker) SetLoopCount(count int32) {
	w.w.SetLoopCount(count)
}

func (w *Worker) SetParallelism(parallelism int32) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	atomic.StoreInt32(&w.parallelism, parallelism)

	w.w.SetParallelism(parallelism)
}

func (w *Worker) AddParallelism(parallelism int32) {
	w.SetParallelism(atomic.LoadInt32(&w.parallelism) + parallelism)
}
