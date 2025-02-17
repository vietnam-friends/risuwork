package workerx

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/isucon/isucandar/worker"
)

type WorkerOption func(*Worker) error

func WithLoopCount(count int32) WorkerOption {
	return func(w *Worker) error {
		w.opts = append(w.opts, worker.WithLoopCount(count))
		return nil
	}
}

func WithInfinityLoop() WorkerOption {
	return func(w *Worker) error {
		w.opts = append(w.opts, worker.WithInfinityLoop())
		return nil
	}
}

func WithUnlimitedParallelism() WorkerOption {
	return func(w *Worker) error {
		w.parallelism = -1
		w.opts = append(w.opts, worker.WithUnlimitedParallelism())
		return nil
	}
}

func WithParallelismUpdater(f func(int) int32) WorkerOption {
	return func(w *Worker) error {
		w.parallelismUpdater = f
		initParallelism := f(0)
		w.parallelism = initParallelism
		w.opts = append(w.opts, worker.WithMaxParallelism(initParallelism))
		slog.Debug(fmt.Sprintf("parallelism updater set to %d", f(0)), "level", 0, "parallelism", f(0))
		return nil
	}
}

func WithConstantParallelism(n int) WorkerOption {
	return WithParallelismUpdater(func(lv int) int32 {
		return int32(n)
	})
}

func WithLinearParallelismUpdate(n int) WorkerOption {
	return WithParallelismUpdater(func(lv int) int32 {
		return int32(n) * int32(lv)
	})
}

func WithExponentialParallelismUpdate(n int) WorkerOption {
	return WithParallelismUpdater(func(lv int) int32 {
		return int32(math.Pow(float64(n), float64(lv)))
	})
}

func WithFixedDelay(delay time.Duration) WorkerOption {
	return func(w *Worker) error {
		old := w.workFunc
		w.workFunc = func(ctx context.Context, i int) {
			old(ctx, i)
			ticker := time.NewTicker(delay)
			defer ticker.Stop()
			select {
			case <-ticker.C:
			case <-ctx.Done():
			}
		}
		return nil
	}
}

func WithFixedInterval(interval time.Duration) WorkerOption {
	return func(w *Worker) error {
		old := w.workFunc
		w.workFunc = func(ctx context.Context, i int) {
			go old(ctx, i)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			select {
			case <-ticker.C:
			case <-ctx.Done():
			}
		}
		return nil
	}
}
