package internal

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/samber/lo"
	"github.com/zckevin/tcp-link-inspect/internal/types"
	"golang.org/x/exp/constraints"
)

func MapAllConcurrently[KeyT any, ResultT any](
	oldctx context.Context,
	keys []KeyT,
	callback func(ctx context.Context, key KeyT) (ResultT, error),
) ([]ResultT, error) {
	ctx, cancel := context.WithCancel(oldctx)
	defer cancel()

	N := len(keys)
	results := make([]ResultT, N)
	errCh := make(chan error, N)
	fn := func(i int, key KeyT) {
		result, err := callback(ctx, key)
		errCh <- err
		if err == nil {
			results[i] = result
		}
	}
	for index, ip := range keys {
		go fn(index, ip)
	}
	for i := 0; i < N; i++ {
		err := <-errCh
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func MapAllConcurrentlyAllSettled[KeyT any, ResultT any](
	oldctx context.Context,
	keys []KeyT,
	callback func(ctx context.Context, key KeyT) (ResultT, error),
) []types.Result[ResultT] {
	ctx, cancel := context.WithCancel(oldctx)
	defer cancel()

	N := len(keys)
	results := make([]types.Result[ResultT], N)
	var wg sync.WaitGroup
	wg.Add(N)
	fn := func(i int, key KeyT) {
		defer wg.Done()
		results[i] = types.AsResult(callback(ctx, key))
	}
	for index, ip := range keys {
		go fn(index, ip)
	}
	wg.Wait()
	return results
}

func PromiseAll(
	oldctx context.Context,
	callbacks []func(ctx context.Context) error,
) error {
	ctx, cancel := context.WithCancel(oldctx)
	defer cancel()

	N := len(callbacks)
	errCh := make(chan error, N)
	fn := func(i int) {
		err := callbacks[i](ctx)
		errCh <- err
	}
	for i := 0; i < N; i++ {
		go fn(i)
	}
	for i := 0; i < N; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func UnwrapAll[T any](results []types.Result[*T]) []*T {
	return lo.Map(
		lo.Filter(results, func(result types.Result[*T], _ int) bool {
			return result.IsOk() && (result.Unwrap() != nil)
		}),
		func(result types.Result[*T], _ int) *T {
			return result.Unwrap()
		})
}

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Abs[T constraints.Signed](a T) T {
	if a < 0 {
		return -a
	}
	return a
}

func SpawnPprofServer() {
	// if !pprofServerStarted.CompareAndSwap(false, true) {
	// 	return
	// }
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			panic(err)
		}
	}()
}
